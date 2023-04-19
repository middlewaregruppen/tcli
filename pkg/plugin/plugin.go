package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	username           string
	password           string
	namespace          string
	cluster            string
	server             string
	insecureSkipVerify bool
	f                  *os.File
}

var pluginName = "kubectl-vsphere"

func IsCommandAvailable() error {
	return exec.Command("/bin/sh", "-c", "command -v "+pluginName).Run()
}

func New() *Client {
	return &Client{
		username:           "",
		password:           "",
		cluster:            "",
		namespace:          "",
		server:             "",
		insecureSkipVerify: false,
	}
}

func (p *Client) Username(u string) *Client {
	p.username = u
	return p
}

func (p *Client) Password(pwd string) *Client {
	p.password = pwd
	return p
}

func (p *Client) Insecure(t bool) *Client {
	p.insecureSkipVerify = t
	return p
}

func (p *Client) Server(s string) *Client {
	p.server = s
	return p
}

func (p *Client) Namespace(ns string) *Client {
	p.namespace = ns
	return p
}

func (p *Client) Cluster(c string) *Client {
	p.cluster = c
	return p
}

// KUBECTL_VSPHERE_PASSWORD='Gde$pj8X!V$Y' kubectl vsphere login \
//   --server=sko-vcf-w01-sc01-api \
//   --insecure-skip-tls-verify \
//   --tanzu-kubernetes-cluster-namespace sko-vcf-w01-shared01-dev --tanzu-kubernetes-cluster-name iris-dev

func (p *Client) Clusters(ns string) (string, error) {
	args := []string{"login", "--vsphere-username", p.username, "--server", p.server, "--tanzu-kubernetes-cluster-namespace", p.namespace}
	if p.insecureSkipVerify {
		args = append(args, "--insecure-skip-tls-verify")
	}

	// Create temporary kubeconfig file
	f, err := os.CreateTemp("", "tanzu-login-") // in Go version older than 1.17 you can use ioutil.TempFile
	if err != nil {
		return "", err
	}
	fstat, err := f.Stat()
	if err != nil {
		return "", err
	}
	defer f.Close()
	p.f = f

	var out strings.Builder

	cmd := exec.Command(pluginName, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECTL_VSPHERE_PASSWORD=%s", p.password))
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", fstat.Name()))
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(out.String(), err)
	}
	return out.String(), nil
}

func (p *Client) Clean() error {
	return nil
}
