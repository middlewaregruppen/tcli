package login

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"syscall"

	"github.com/middlewaregruppen/tcli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	tanzuNamespace string
	silent         bool
)

func NewCmdLogin() *cobra.Command {
	c := &cobra.Command{
		Use:   "login [CLUSTER...]",
		Args:  cobra.MinimumNArgs(0),
		Short: "Authenticate user with Tanzu namespaces and clusters",
		Long: `Authenticate user with Tanzu namespaces and clusters
Examples:
	# Login to the supervisor cluster
	tcli -s SERVER -u USER -p PASSWORD login

	# Flags can be prefixed with TCLI_ and therefore omitted from the command line
	export TCLI_SERVER=https://supervisor.local
	export TCLI_USERNAME=bob
	export TCLI_PASSWORD=mypassword
	tcli login

	# Login to a tanzu cluster
	tcli login CLUSTER

	# Login to multiple tanzu clusters in one go
	tcli login CLUSTER1 CLUSTER2 CLUSTER3 ...

	# Login to tanzu clusters in the same namespace
	tcli login CLUSTER1 CLUSTER2 -n NAMESPACE

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Prompt for password on stdin if it wasn't provided via flag or env
			if len(viper.GetString("password")) == 0 {
				fmt.Printf("Password:")
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return err
				}
				if err := cmd.Flags().Set("password", string(bytePassword)); err != nil {
					return err
				}
				fmt.Printf("\n")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()

			tanzuServer := viper.GetString("server")
			tanzuUsername := viper.GetString("username")
			tanzuPassword := viper.GetString("password")
			tanzuNamespace := viper.GetString("namespace")
			insecureSkipVerify := viper.GetBool("insecure")
			kubeconfig := viper.GetString("kubeconfig")

			u, err := url.Parse(tanzuServer)
			if err != nil {
				return fmt.Errorf("parsing server URL: %w", err)
			}

			// Supervisor server speaks the WCP API on port 443, but the
			// kubeconfig cluster entry uses the k8s API on port 6443. Build
			// the k8s API URL from the already-parsed host to avoid doubling
			// the port if the user supplied tanzuServer with an explicit port.
			supervisorK8sServer := fmt.Sprintf("https://%s:6443", u.Hostname())

			c, err := client.New(
				tanzuServer,
				client.WithLogger(slog.Default()),
				client.WithCredentials(client.BasicCredentials(tanzuUsername, tanzuPassword)),
				client.WithInsecure(insecureSkipVerify),
			)
			if err != nil {
				return err
			}

			sess, err := c.Login(ctx, tanzuUsername, tanzuPassword)
			if err != nil {
				return err
			}

			ns, err := c.Namespaces(ctx)
			if err != nil {
				return err
			}

			// Build the supervisor cluster entry for kubeconfig
			supervisorCluster := api.NewCluster()
			supervisorCluster.InsecureSkipTLSVerify = insecureSkipVerify
			supervisorCluster.Server = supervisorK8sServer

			authName := fmt.Sprintf("wcp:%s:%s", u.Host, tanzuUsername)
			auth := api.NewAuthInfo()
			auth.Token = sess.SessionID

			kubectx := api.NewContext()
			kubectx.Cluster = u.Host
			kubectx.AuthInfo = authName
			if len(ns) > 0 {
				kubectx.Namespace = ns[len(ns)-1].Namespace
			}

			// Load kubeconfig once; update in memory, write once at the end
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return fmt.Errorf("loading kubeconfig: %w", err)
			}
			conf.Clusters[u.Host] = supervisorCluster
			conf.AuthInfos[authName] = auth
			conf.Contexts[u.Host] = kubectx
			conf.CurrentContext = u.Host

			if !silent {
				fmt.Printf("You have access to following %d namespaces:\n", len(ns))
				for _, n := range ns {
					fmt.Println(n.Namespace)
				}
			}

			// Login to each requested workload cluster, updating conf in memory
			for _, tanzuCluster := range args {
				res, err := c.LoginCluster(ctx, tanzuCluster, tanzuNamespace)
				if err != nil {
					if errors.Is(err, client.ErrClusterNotFound) {
						return fmt.Errorf("cluster %q not found", tanzuCluster)
					}
					return err
				}

				caCertData, err := base64.StdEncoding.DecodeString(res.GuestClusterCa)
				if err != nil {
					return fmt.Errorf("decoding CA cert for cluster %q: %w", tanzuCluster, err)
				}

				wlCluster := api.NewCluster()
				wlCluster.CertificateAuthorityData = caCertData
				wlCluster.Server = fmt.Sprintf("https://%s:6443", res.GuestClusterServer)

				wlAuthName := fmt.Sprintf("wcp:%s:%s", res.GuestClusterServer, tanzuUsername)
				wlAuth := api.NewAuthInfo()
				wlAuth.Token = res.SessionID

				wlCtx := api.NewContext()
				wlCtx.Cluster = res.GuestClusterServer
				wlCtx.AuthInfo = wlAuthName

				// Propagate the namespace into the supervisor context so that
				// subsequent commands that read --namespace from kubeconfig work
				// without requiring the flag explicitly.
				if _, ok := conf.Contexts[u.Host]; ok {
					conf.Contexts[u.Host].Namespace = tanzuNamespace
				}

				conf.Clusters[res.GuestClusterServer] = wlCluster
				conf.AuthInfos[wlAuthName] = wlAuth
				conf.Contexts[tanzuCluster] = wlCtx
				conf.CurrentContext = tanzuCluster

				fmt.Printf("Successfully logged into cluster %s\n", tanzuCluster)
			}

			// Single write after all in-memory updates are done
			if err := clientcmd.WriteToFile(*conf, kubeconfig); err != nil {
				return fmt.Errorf("writing kubeconfig: %w", err)
			}

			return nil
		},
	}
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace in which the Tanzu Kubernetes cluster resides.")
	c.Flags().BoolVar(&silent, "silent", false, "Silent mode - suppress output")
	return c
}
