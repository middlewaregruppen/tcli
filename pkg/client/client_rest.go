package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrClusterNotFound        = errors.New("cluster not found")
	_                  Client = &RestClient{}

	PathWCPWorkloads            string = "/wcp/workloads"
	PathWCPLogin                string = "/wcp/login"
	PathTanzuKubernetesClusters string = "/apis/run.tanzu.vmware.com/v1alpha2/namespaces/%s/tanzukubernetesclusters"
	PathTanzuKubernetesCluster  string = "/apis/run.tanzu.vmware.com/v1alpha2/namespaces/%s/tanzukubernetesclusters/%s"
	PathTanzuKubernetesReleases string = "/apis/run.tanzu.vmware.com/v1alpha2/tanzukubernetesreleases"
	PathTanzuKubernetesAddons   string = "/apis/run.tanzu.vmware.com/v1alpha2/tanzukubernetesaddons"
)

type RestClient struct {
	uri        *url.URL
	httpClient *http.Client
	auth       Credentials
	Token      string
	logger     *slog.Logger
}

type Credentials interface {
	Apply(*http.Request) error
}

type basicCredentials struct {
	username string
	password string
}

func (c *basicCredentials) Apply(r *http.Request) error {
	r.SetBasicAuth(c.username, c.password)
	return nil
}

type tokenCredentials string

func (c *tokenCredentials) Apply(r *http.Request) error {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", *c))
	return nil
}

func TokenCredentials(token string) Credentials {
	t := tokenCredentials(token)
	return &t
}

func BasicCredentials(username, password string) Credentials {
	return &basicCredentials{
		username: username,
		password: password,
	}
}

type Option func(*RestClient)

type LoginResponse struct {
	SessionID string `json:"session_id,omitempty"`
}

type LoginClusterResponse struct {
	LoginResponse
	GuestClusterServer string `json:"guest_cluster_server,omitempty"`
	GuestClusterCa     string `json:"guest_cluster_ca,omitempty"`
}

type Namespace struct {
	Namespace                string   `json:"namespace,omitempty"`
	MasterHost               string   `json:"master_host,omitempty"`
	ConrolPlaneAPIServerPort string   `json:"conrol_plane_api_server_port,omitempty"`
	ControlPlaneDNSNames     []string `json:"control_plane_dns_names,omitempty"`
}

func WithClient(c *http.Client) Option {
	return func(r *RestClient) {
		r.httpClient = c
	}
}

func WithInsecure(insecure bool) Option {
	return func(rc *RestClient) {
		if rc.httpClient == nil {
			rc.httpClient = &http.Client{}
		}

		baseTransport := rc.httpClient.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}

		transport := baseTransport.(*http.Transport).Clone()
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = insecure
		rc.httpClient.Transport = transport
	}
}

func WithCredentials(creds Credentials) Option {
	return func(rc *RestClient) {
		rc.auth = creds
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(rc *RestClient) {
		rc.logger = l
	}
}

func (r *RestClient) SetToken(t string) *RestClient {
	r.Token = t
	return r
}

// DoRequest applies options and then performs the http request
func (r *RestClient) DoRequest(req *http.Request) (*http.Response, error) {
	if r.auth != nil {
		if err := r.auth.Apply(req); err != nil {
			r.logger.Debug("error applying credentials to request",
				"method", req.Method,
				"url", req.URL.String(),
				"auth_method", &r.auth,
				"error", err,
			)
			return nil, err
		}
	}

	req.Header.Add("Content-Type", "application/json")
	start := time.Now()

	res, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Debug("http request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"duration", time.Since(start),
			"error", err,
		)
		return nil, err
	}

	r.logger.Debug("http request executed",
		"method", req.Method,
		"url", req.URL.String(),
		"duration", time.Since(start),
		"error", err,
		"res_content_type", res.Header.Values("Content-Type"),
		"req_content_type", req.Header.Values("Content-Type"),
	)
	return res, nil
}

// getRequestURI builds an URI for the given path. It uses the base URI when RestClient was created with [New]
func (r *RestClient) getRequestURI(path string) (*url.URL, error) {
	newPath, err := url.JoinPath(r.uri.Path, path)
	if err != nil {
		return nil, err
	}
	newURL := *r.uri
	newURL.Path = newPath
	r.logger.Debug("built request uri",
		"path", path,
		"uri", newURL.String(),
	)
	return &newURL, nil
}

func (r *RestClient) Namespaces(ctx context.Context) ([]Namespace, error) {
	u, err := r.getRequestURI(PathWCPWorkloads)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var namespaces []Namespace
	err = json.Unmarshal(body, &namespaces)
	if err != nil {
		return nil, err
	}

	// Sort namespaces by name
	sort.SliceStable(namespaces, func(i, j int) bool {
		return namespaces[i].Namespace < namespaces[j].Namespace
	})

	return namespaces, nil
}

func (r *RestClient) Clusters(ctx context.Context, ns string) (*v1.Table, error) {
	if len(ns) == 0 {
		ns = "default"
	}

	u, err := r.getRequestURI(fmt.Sprintf(PathTanzuKubernetesClusters, ns))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json;as=Table;g=meta.k8s.io;v=v1")

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var clusterlist v1.Table
	err = json.Unmarshal(body, &clusterlist)
	if err != nil {
		return nil, err
	}

	return &clusterlist, nil
}

func (r *RestClient) ReleasesTable(ctx context.Context) (*v1.Table, error) {
	u, err := r.getRequestURI(PathTanzuKubernetesReleases)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json;as=Table;g=meta.k8s.io;v=v1")

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var releases v1.Table
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, err
	}

	return &releases, nil
}

func (r *RestClient) AddonsTable(ctx context.Context) (*v1.Table, error) {
	u, err := r.getRequestURI(PathTanzuKubernetesAddons)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json;as=Table;g=meta.k8s.io;v=v1")

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var addons v1.Table
	err = json.Unmarshal(body, &addons)
	if err != nil {
		return nil, err
	}

	return &addons, nil
}

func (r *RestClient) Releases(ctx context.Context) (*v1alpha2.TanzuKubernetesReleaseList, error) {
	u, err := r.getRequestURI(PathTanzuKubernetesReleases)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json;as=Table;g=meta.k8s.io;v=v1")

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var releases v1alpha2.TanzuKubernetesReleaseList
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, err
	}

	return &releases, nil
}

func (r *RestClient) Cluster(ctx context.Context, ns, name string) (*v1alpha2.TanzuKubernetesCluster, error) {
	u, err := r.getRequestURI(fmt.Sprintf(PathTanzuKubernetesCluster, ns, name))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var cluster v1alpha2.TanzuKubernetesCluster
	err = json.Unmarshal(body, &cluster)
	if err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (r *RestClient) Login(ctx context.Context, u, p string) (*LoginResponse, error) {
	uri, err := r.getRequestURI(PathWCPLogin)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}
	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var login LoginResponse
	err = json.Unmarshal(body, &login)
	if err != nil {
		return nil, err
	}

	return &login, nil
}

func (r *RestClient) LoginCluster(ctx context.Context, cluster, namespace string) (*LoginClusterResponse, error) {
	data := fmt.Sprintf("{\"guest_cluster_name\":\"%s\"}", cluster)
	if len(namespace) > 0 {
		data = fmt.Sprintf("{\"guest_cluster_name\":\"%s\", \"guest_cluster_namespace\":\"%s\"}", cluster, namespace)
	}

	uri, err := r.getRequestURI(PathWCPLogin)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}

	resp, err := r.DoRequest(req)
	if err != nil {
		return nil, err
	}

	body, err := r.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var login LoginClusterResponse
	err = json.Unmarshal(body, &login)
	if err != nil {
		return nil, err
	}

	// An 'guest_cluster_server' in response means a not-found error
	if len(login.GuestClusterServer) == 0 {
		return nil, ErrClusterNotFound
	}
	return &login, nil
}

func (r *RestClient) handleResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Debug("error reading response body",
			"error", err,
		)
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	r.logger.Debug("read http response body",
		"status_code", resp.StatusCode,
		"length", len(body),
		"content_type", resp.Header.Values("Content-Type"),
	)

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return nil, errors.New(string(body))
	}
	return body, nil
}

func New(baseURI string, opts ...Option) (Client, error) {
	u, err := url.ParseRequestURI(baseURI)
	if err != nil {
		return nil, err
	}
	c := &RestClient{
		uri:        u,
		httpClient: http.DefaultClient,
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}
