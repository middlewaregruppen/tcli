package logout

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdLogout() *cobra.Command {
	c := &cobra.Command{
		Use:   "logout",
		Short: "Logout user and remove credentials",
		Long:  "",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			tanzuUsername := viper.GetString("username")
			kubeconfig := viper.GetString("kubeconfig")

			// Read kubeconfig from file
			conf, err := clientcmd.LoadFromFile(kubeconfig)
			if err != nil {
				return err
			}

			for k, v := range conf.Contexts {
				userSplit := strings.Split(v.AuthInfo, ":")
				if len(userSplit) == 3 {
					if userSplit[0] == "wcp" && userSplit[2] == tanzuUsername {
						delete(conf.Clusters, v.Cluster)
						delete(conf.AuthInfos, v.AuthInfo)
						delete(conf.Contexts, k)
					}
				}
			}

			// Write back to kubeconfig
			err = clientcmd.WriteToFile(*conf, kubeconfig)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return c
}
