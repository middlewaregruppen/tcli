package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
)

type RestClient struct {
	u        *url.URL
	c        *http.Client
	username string
	password string
	Token    string
}

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

func New(baseUrl string) (*RestClient, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &RestClient{
		u: u,
		c: http.DefaultClient,
	}, nil
}

func (r *RestClient) SetInsecure(t bool) *RestClient {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return r
}

func (r *RestClient) SetToken(t string) *RestClient {
	r.Token = t
	return r
}

func (r *RestClient) Namespaces() ([]Namespace, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/wcp/workloads", r.u.String()), nil)
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.username, r.password))))},
	}
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var namespaces []Namespace
	err = json.Unmarshal(body, &namespaces)
	if err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (r *RestClient) Clusters(ns string) (*v1alpha2.TanzuKubernetesClusterList, error) {
	if len(ns) == 0 {
		ns = "default"
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s:6443/apis/run.tanzu.vmware.com/v1alpha2/namespaces/%s/tanzukubernetesclusters?limit=500", r.u.String(), ns), nil)
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", r.Token)},
	}
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := handleResponse(resp)
	if err != nil {
		return nil, err
	}
	var clusterlist v1alpha2.TanzuKubernetesClusterList
	err = json.Unmarshal(body, &clusterlist)
	if err != nil {
		return nil, err
	}
	return &clusterlist, nil
}

func (r *RestClient) Login(u, p string) error {

	r.username = u
	r.password = p
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/wcp/login", r.u.String()), nil)
	if err != nil {
		return err
	}
	req.Header = map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", u, p))))},
	}
	resp, err := r.c.Do(req)
	if err != nil {
		return err
	}
	body, err := handleResponse(resp)
	if err != nil {
		return err
	}

	var login LoginResponse
	err = json.Unmarshal(body, &login)
	if err != nil {
		return err
	}
	r.Token = login.SessionID
	return nil
}

func (r *RestClient) LoginCluster(cluster, namespace string) (*LoginClusterResponse, error) {
	data := fmt.Sprintf("{\"guest_cluster_name\":\"%s\", \"guest_cluster_namespace\":\"%s\"}", cluster, namespace)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/wcp/login", r.u.String()), bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.username, r.password))))},
	}
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := handleResponse(resp)
	if err != nil {
		return nil, err
	}

	var login LoginClusterResponse
	err = json.Unmarshal(body, &login)
	if err != nil {
		return nil, err
	}
	return &login, nil
}

func handleResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return nil, errors.New(string(body))
	}
	return body, nil
}
