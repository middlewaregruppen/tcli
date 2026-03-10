package use

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/client-go/tools/clientcmd"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig := viper.GetString("kubeconfig")
			namespace := args[0]

			// Read kubeconfig from file
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return fmt.Errorf("loading kubeconfig: %w", err)
			}

			// Update namespace in current context
			currentCtx := conf.CurrentContext
			if _, ok := conf.Contexts[currentCtx]; ok {
				conf.Contexts[currentCtx].Namespace = namespace
			}

			// Write back to kubeconfig
			if err := clientcmd.WriteToFile(*conf, kubeconfig); err != nil {
				return fmt.Errorf("writing kubeconfig: %w", err)
			}

			fmt.Printf("Namespace set to %q in context %q\n", namespace, currentCtx)
			return nil
		},
	}
	return c
}
