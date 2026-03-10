package logout

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdLogout() *cobra.Command {
	c := &cobra.Command{
		Use:   "logout",
		Short: "Logout user and remove credentials",
		Long: `Logout user and remove all WCP credentials from the kubeconfig.

All contexts, clusters, and authinfos that were written by "tcli login" for
the current user are removed from the kubeconfig file.

Examples:
	# Logout the current user
	tcli -s SERVER -u USER logout

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tanzuUsername := viper.GetString("username")
			kubeconfig := viper.GetString("kubeconfig")

			// Read kubeconfig from file
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return fmt.Errorf("loading kubeconfig: %w", err)
			}

			removed := 0
			for k, v := range conf.Contexts {
				userSplit := strings.Split(v.AuthInfo, ":")
				if len(userSplit) == 3 && userSplit[0] == "wcp" && userSplit[2] == tanzuUsername {
					delete(conf.Clusters, v.Cluster)
					delete(conf.AuthInfos, v.AuthInfo)
					delete(conf.Contexts, k)
					removed++
				}
			}

			if removed == 0 {
				fmt.Printf("No credentials found for user %q — nothing to remove.\n", tanzuUsername)
				return nil
			}

			// Write back to kubeconfig
			if err := clientcmd.WriteToFile(*conf, kubeconfig); err != nil {
				return fmt.Errorf("writing kubeconfig: %w", err)
			}

			fmt.Printf("Removed %d context(s) for user %q.\n", removed, tanzuUsername)
			return nil
		},
	}
	return c
}
