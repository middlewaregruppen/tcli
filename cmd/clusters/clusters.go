package clusters

import (
	"errors"
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

			// Create rest client
			c, err := client.New(tanzuServer)
			if err != nil {
				return err
			}

			// Find credentials from kubeconfig context
			contextName := u.Host
			if _, ok := conf.Contexts[contextName]; !ok {
				return errors.New("credentials missing! Please run 'tcli login' to authenticate")
			}

			// AuthInfo name is whatever is set in the context. However it can be overriden with the --username flag
			authName := fmt.Sprintf("wcp:%s:%s", u.Host, conf.Contexts[contextName].AuthInfo)
			if len(tanzuUsername) > 0 {
				authName = fmt.Sprintf("wcp:%s:%s", u.Host, tanzuUsername)
			}

			// Check if the AuthInfo object exists
			if _, ok := conf.AuthInfos[authName]; !ok {
				return errors.New("credentials missing! Please run 'tcli login' to authenticate")
			}
			c.SetToken(conf.AuthInfos[authName].Token)
			c.SetInsecure(insecureSkipVerify)

			// Check if there is a namespace set in the context that we can use so that we don't have to specify the --namespace flag
			if _, ok := conf.Contexts[contextName]; ok && len(tanzuNamespace) == 0 {
				tanzuNamespace = conf.Contexts[contextName].Namespace
			}

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
