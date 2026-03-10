package list

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/middlewaregruppen/tcli/cmd/internal/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/middlewaregruppen/tcli/pkg/client"
)

var tanzuNamespace string

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

	# List addons
	tcli list addons

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()

			tanzuServer := viper.GetString("server")
			tanzuUsername := viper.GetString("username")
			tanzuPassword := viper.GetString("password")
			insecureSkipVerify := viper.GetBool("insecure")
			kubeconfig := viper.GetString("kubeconfig")

			c, contextNamespace, err := auth.ClientFromKubeconfig(tanzuServer, kubeconfig, tanzuUsername, insecureSkipVerify)
			if err != nil {
				return err
			}

			// If --namespace was not given, fall back to the namespace stored in the kubeconfig context
			if len(tanzuNamespace) == 0 {
				tanzuNamespace = contextNamespace
			}

			switch strings.ToLower(args[0]) {
			case "namespaces", "ns":
				return listNamespaces(ctx, tanzuServer, tanzuUsername, tanzuPassword, insecureSkipVerify)
			case "clusters", "clu", "tkc":
				return listClusters(ctx, c, tanzuNamespace)
			case "releases", "rel", "tkr":
				return listReleases(ctx, c)
			case "addons", "tka":
				return listAddons(ctx, c)
			default:
				return fmt.Errorf("%q is not a valid resource", args[0])
			}
		},
	}
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace in which the Tanzu Kubernetes cluster resides.")
	return c
}

func listClusters(ctx context.Context, c client.Client, ns string) error {
	objs, err := c.Clusters(ctx, ns)
	if err != nil {
		return err
	}
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	return printer.PrintObj(objs, os.Stdout)
}

func listReleases(ctx context.Context, c client.Client) error {
	objs, err := c.ReleasesTable(ctx)
	if err != nil {
		return err
	}
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	return printer.PrintObj(objs, os.Stdout)
}

func listNamespaces(ctx context.Context, server, username, password string, insecure bool) error {
	c, err := client.New(server, client.WithCredentials(client.BasicCredentials(username, password)), client.WithInsecure(insecure))
	if err != nil {
		return err
	}

	nsList, err := c.Namespaces(ctx)
	if err != nil {
		return err
	}
	for _, n := range nsList {
		fmt.Println(n.Namespace)
	}
	return nil
}

func listAddons(ctx context.Context, c client.Client) error {
	objs, err := c.AddonsTable(ctx)
	if err != nil {
		return err
	}
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	return printer.PrintObj(objs, os.Stdout)
}
