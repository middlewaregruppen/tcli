package clusters

import (
	"fmt"
	"net/url"

	"github.com/middlewaregruppen/tcli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	tanzuNamespace string
)

func NewCmdClusters() *cobra.Command {
	c := &cobra.Command{
		Use:   "clusters",
		Short: "List clusters within a Tanzu namespace",
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
			insecureSkipVerify := viper.GetBool("insecure")
			kubeconfig := viper.GetString("kubeconfig")

			u, err := url.Parse(tanzuServer)
			if err != nil {
				return err
			}

			// Read kubeconfig from file
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return err
			}

			c, err := client.New(tanzuServer)
			if err != nil {
				return err
			}

			authName := fmt.Sprintf("wcp:%s:%s", u.Host, tanzuUsername)
			c.SetInsecure(insecureSkipVerify)
			c.SetToken(conf.AuthInfos[authName].Token)

			clusterlist, err := c.Clusters(tanzuNamespace)
			if err != nil {
				return err
			}

			fmt.Printf("%d clusters in %s:\n", len(clusterlist.Items), tanzuNamespace)
			for _, n := range clusterlist.Items {
				fmt.Println(n.Name)
			}
			return nil
		},
	}
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace in which the Tanzu Kubernetes cluster resides.")
	return c
}
