// Package auth provides shared helpers for resolving credentials from a
// kubeconfig file and constructing an authenticated API client.
package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/middlewaregruppen/tcli/pkg/client"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ErrNotAuthenticated is returned when the expected kubeconfig context or
// authinfo is missing — i.e. the user has not run "tcli login" yet.
var ErrNotAuthenticated = errors.New("credentials missing! Please run 'tcli login' to authenticate")

// ClientFromKubeconfig loads the kubeconfig at kubeconfigPath, resolves the
// stored session token for the given server, and returns a ready-to-use
// client.Client together with the resolved namespace from the context.
//
// If username is non-empty it overrides the username stored in the context
// when constructing the authinfo key.
func ClientFromKubeconfig(server, kubeconfigPath, username string, insecure bool) (client.Client, string, error) {
	u, err := url.Parse(server)
	if err != nil {
		return nil, "", fmt.Errorf("parsing server URL: %w", err)
	}

	conf, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("loading kubeconfig: %w", err)
	}

	token, namespace, err := TokenFromConfig(conf, u.Host, username)
	if err != nil {
		return nil, "", err
	}

	c, err := client.New(
		server,
		client.WithLogger(slog.Default()),
		client.WithCredentials(client.TokenCredentials(token)),
		client.WithInsecure(insecure),
	)
	if err != nil {
		return nil, "", fmt.Errorf("creating client: %w", err)
	}

	return c, namespace, nil
}

// TokenFromConfig resolves the session token and context namespace from an
// already-loaded kubeconfig, given the supervisor host (u.Host) and an
// optional username override.
func TokenFromConfig(conf *clientcmdapi.Config, host, username string) (token, namespace string, err error) {
	ctx, ok := conf.Contexts[host]
	if !ok {
		return "", "", ErrNotAuthenticated
	}

	authName := fmt.Sprintf("wcp:%s:%s", host, ctx.AuthInfo)
	if len(username) > 0 {
		authName = fmt.Sprintf("wcp:%s:%s", host, username)
	}

	authInfo, ok := conf.AuthInfos[authName]
	if !ok {
		return "", "", ErrNotAuthenticated
	}

	return authInfo.Token, ctx.Namespace, nil
}
