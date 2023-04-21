package login

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/middlewaregruppen/tcli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	tanzuCluster   string
	tanzuNamespace string
)

func NewCmdLogin() *cobra.Command {
	c := &cobra.Command{
		Use:   "login",
		Short: "Authenticate user with Tanzu namespaces and clusters",
		Long:  "",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			tanzuServer := viper.GetString("server")
			tanzuUsername := viper.GetString("username")
			tanzuPassword := viper.GetString("password")
			tanzuCluster := viper.GetString("cluster")
			tanzuNamespace := viper.GetString("namespace")
			insecureSkipVerify := viper.GetBool("insecure")
			kubeconfig := viper.GetString("kubeconfig")

			u, err := url.Parse(tanzuServer)
			if err != nil {
				return err
			}

			c, err := client.New(tanzuServer)
			if err != nil {
				return err
			}
			c.SetInsecure(insecureSkipVerify)
			err = c.Login(tanzuUsername, tanzuPassword)
			if err != nil {
				return err
			}

			ns, err := c.Namespaces()
			if err != nil {
				return err
			}

			// Define the new cluster to which we have logged in to
			cluster := api.NewCluster()
			cluster.InsecureSkipTLSVerify = true
			cluster.Server = fmt.Sprintf("%s:6443", tanzuServer)

			authName := fmt.Sprintf("wcp:%s:%s", u.Host, tanzuUsername)
			auth := api.NewAuthInfo()
			auth.Token = c.Token

			context := api.NewContext()
			context.Cluster = u.Host
			context.AuthInfo = authName

			// Read kubeconfig from file
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return err
			}
			conf.Clusters[u.Host] = cluster
			conf.AuthInfos[authName] = auth
			conf.Contexts[u.Host] = context

			// Write back to kubeconfig
			err = clientcmd.WriteToFile(*conf, kubeconfig)
			if err != nil {
				return err
			}

			fmt.Printf("You have access to following %d namespaces:\n", len(ns))
			for _, n := range ns {
				fmt.Println(n.Namespace)
			}

			// Login to cluster if both flags are present
			if len(tanzuCluster) > 0 && len(tanzuNamespace) > 0 {
				res, err := c.LoginCluster(tanzuCluster, tanzuNamespace)
				if err != nil {
					return err
				}
				caCertData, err := base64.StdEncoding.DecodeString(res.GuestClusterCa)
				if err != nil {
					return err
				}
				cluster := api.NewCluster()
				cluster.CertificateAuthorityData = caCertData
				cluster.Server = fmt.Sprintf("https://%s:6443", res.GuestClusterServer)
				authName := fmt.Sprintf("wcp:%s:%s", res.GuestClusterServer, tanzuUsername)
				auth := api.NewAuthInfo()
				auth.Token = res.SessionID
				context := api.NewContext()
				context.Cluster = res.GuestClusterServer
				context.AuthInfo = authName

				conf, err := clientcmd.LoadFromFile(kubeconfig)
				if err != nil {
					return err
				}

				// Update namespace field in context for supervisor server.
				// This allows us to run commands that require the --namespace flag without having to providing the flag
				if _, ok := conf.Contexts[u.Host]; ok {
					conf.Contexts[u.Host].Namespace = tanzuNamespace
				}

				conf.Clusters[res.GuestClusterServer] = cluster
				conf.AuthInfos[authName] = auth
				conf.Contexts[tanzuCluster] = context
				conf.CurrentContext = tanzuCluster

				// Write back to kubeconfig
				err = clientcmd.WriteToFile(*conf, kubeconfig)
				if err != nil {
					return err
				}

			}

			return nil
		},
	}
	c.Flags().StringVarP(&tanzuCluster, "cluster", "c", "", "Name of the Tanzu Kubernetes Cluster to authenticate against.")
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace in which the Tanzu Kubernetes cluster resides.")
	return c
}
