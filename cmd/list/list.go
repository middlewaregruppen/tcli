package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/middlewaregruppen/tcli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/tools/clientcmd"
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
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()

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

			token := conf.AuthInfos[authName].Token

			// Create rest client
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			c, err := client.New(tanzuServer, client.WithLogger(logger), client.WithCredentials(client.TokenCredentials(token)), client.WithInsecure(insecureSkipVerify))
			if err != nil {
				return err
			}

			if _, ok := conf.Contexts[contextName]; ok && len(tanzuNamespace) == 0 {
				tanzuNamespace = conf.Contexts[contextName].Namespace
			}

			a := strings.ToLower(args[0])
			switch a {
			case "namespaces", "ns":
				// Read from stdin if password isn't set anywhere
				if len(tanzuPassword) == 0 {
					fmt.Printf("Password:")
					bytePassword, err := term.ReadPassword(int(syscall.Stdin))
					if err != nil {
						return err
					}
					tanzuPassword = string(bytePassword)
					fmt.Printf("\n")
				}
				return listNamespaces(ctx, tanzuServer, tanzuUsername, tanzuPassword, insecureSkipVerify)
			case "clusters", "clu", "tkc":
				return listClusters(ctx, c, tanzuNamespace)
			case "releases", "rel", "tkr":
				return listReleases(ctx, c)
			case "addons", "tka":
				return listAddons(ctx, c)
			default:
				return fmt.Errorf("%s is not a valid resource", a)
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
	err = printer.PrintObj(objs, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}

func listReleases(ctx context.Context, c client.Client) error {
	objs, err := c.ReleasesTable(ctx)
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
	err = printer.PrintObj(objs, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}
