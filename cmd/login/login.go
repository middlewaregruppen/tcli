package login

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/amimof/tanzu-login/pkg/client"
	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	tanzuCluster   string
	tanzuNamespace string
)

func NewCmdLogin() *cobra.Command {
	c := &cobra.Command{
		Use:   "login CLUSTER_NAME",
		Short: "Authenticate user with Tanzu namespaces and clusters",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			tanzuServer, _ := cmd.Flags().GetString("server")
			tanzuUsername, _ := cmd.Flags().GetString("username")
			tanzuPassword, _ := cmd.Flags().GetString("password")
			insecureSkipVerify, _ := cmd.Flags().GetBool("insecure")
			kubeconfig, _ := cmd.Flags().GetString("kubeconfig")

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

				conf.Clusters[res.GuestClusterServer] = cluster
				conf.AuthInfos[authName] = auth
				conf.Contexts[tanzuCluster] = context

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
