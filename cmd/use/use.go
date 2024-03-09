package use

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/client-go/tools/clientcmd"
)

var (
	tanzuNamespace string
)

func NewCmdUse() *cobra.Command {
	c := &cobra.Command{
		Use:   "use NAMESPACE",
		Args:  cobra.ExactArgs(1),
		Short: "Sets the provided namespace in the current context",
		Long: `Sets the provided namespace in the current context 
Examples:
	# Use the "monitoring" namespace
	tcli use monitoring

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			kubeconfig := viper.GetString("kubeconfig")

			// Read kubeconfig from file
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return err
			}

			// Update namespace in current context
			currentCtx := conf.CurrentContext
			if _, ok := conf.Contexts[currentCtx]; ok {
				conf.Contexts[currentCtx].Namespace = args[0]
			}

			// Write back to kubeconfig
			err = clientcmd.WriteToFile(*conf, kubeconfig)
			if err != nil {
				return err
			}

			return nil
		},
	}
	c.Flags().StringVarP(&tanzuNamespace, "namespace", "n", "", "Namespace to use")
	return c
}
