package inspect

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/middlewaregruppen/tcli/cmd/internal/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var tanzuNamespace string

func NewCmdInspect() *cobra.Command {
	c := &cobra.Command{
		Use:   "inspect CLUSTER",
		Short: "Inspect a specific cluster within a namespace",
		Args:  cobra.ExactArgs(1),
		Long: `Inspect a specific cluster within a namespace
Examples:
	# Inspecting will return the raw cluster specification in YAML format
	tcli inspect NAME -n NAMESPACE

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
			defer cancel()

			tanzuCluster := args[0]
			tanzuServer := viper.GetString("server")
			tanzuUsername := viper.GetString("username")
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

			cluster, err := c.Cluster(ctx, tanzuNamespace, tanzuCluster)
			if err != nil {
				return err
			}

			buf := bytes.Buffer{}
			yamlEncoder := yaml.NewEncoder(&buf)
			yamlEncoder.SetIndent(2)
			if err := yamlEncoder.Encode(cluster); err != nil {
				return fmt.Errorf("encoding cluster as YAML: %w", err)
			}
			if _, err := buf.WriteTo(os.Stdout); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}
			return nil
		},
	}
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace in which the Tanzu Kubernetes cluster resides.")
	return c
}
