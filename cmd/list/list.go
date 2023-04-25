package list

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/middlewaregruppen/tcli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	tanzuNamespace string
)

func NewCmdList() *cobra.Command {
	c := &cobra.Command{
		Use:     "list RESOURCE",
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		Short:   "List clusters and namespaces",
		Long: `List clusters and namespaces
Examples:
	# List namespaces
	tcli list namespaces

	# List clusters in a namespace
	tcli list clusters -n NAMESPACE

	# List releases
	tcli list releases

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
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

			a := strings.ToLower(args[0])
			switch a {
			case "namespaces", "ns":
				return listNamespaces(c, tanzuUsername, tanzuPassword)
			case "clusters", "clu", "tkc":
				return listClusters(c, tanzuNamespace)
			case "releases", "rel", "tkr":
				return listReleases(c)
			default:
				return fmt.Errorf("%s is not a valid resource", a)
			}
		},
	}
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace in which the Tanzu Kubernetes cluster resides.")
	return c
}

func listClusters(c *client.RestClient, ns string) error {
	objs, err := c.Clusters(ns)
	if err != nil {
		return err
	}
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	err = printer.PrintObj(objs, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}

func listReleases(c *client.RestClient) error {
	objs, err := c.ReleasesTable()
	if err != nil {
		return err
	}
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	err = printer.PrintObj(objs, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}

func listNamespaces(c *client.RestClient, username, password string) error {
	err := c.Login(username, password)
	if err != nil {
		return err
	}
	nsList, err := c.Namespaces()
	if err != nil {
		return err
	}
	for _, n := range nsList {
		fmt.Println(n.Namespace)
	}
	return nil
}
