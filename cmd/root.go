package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/middlewaregruppen/tcli/cmd/inspect"
	"github.com/middlewaregruppen/tcli/cmd/list"
	"github.com/middlewaregruppen/tcli/cmd/login"
	"github.com/middlewaregruppen/tcli/cmd/logout"
	"github.com/middlewaregruppen/tcli/cmd/version"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tanzuServer        string
	tanzuUsername      string
	tanzuPassword      string
	insecureSkipVerify bool
	verbosity          string
	kubeconfig         string
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TCLI")
}

func NewDefaultCommand() *cobra.Command {
	c := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Use:           "tcli",
		Short:         "A command line tool that simplifies authentication to Tanzu namespaces and clusters",
		Long: `A command line tool that simplifies authentication to Tanzu namespaces and clusters.
	tcli is a simple CLI tool to:
	- Simplify login process over the default vpshere plugin

	Flags can be prefixed with TCLI_ and therefore omitted from the command line
	
	export TCLI_SERVER=https://supervisor.local
	export TCLI_USERNAME=bob
	export TCLI_PASSWORD=mypassword

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.SetOutput(os.Stdout)
			lvl, err := logrus.ParseLevel(verbosity)
			if err != nil {
				return err
			}
			logrus.SetLevel(lvl)

			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Check if kubeconfig exists, create if it doesn't
			if _, err := os.Stat(kubeconfig); errors.Is(err, os.ErrNotExist) {
				_, err = os.OpenFile(kubeconfig, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
				if err != nil {
					return err
				}
				conf := api.NewConfig()
				err = clientcmd.WriteToFile(*conf, kubeconfig)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}
	// Setup flags
	c.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", "info", "number for the log level verbosity (debug, info, warn, error, fatal, panic)")
	c.PersistentFlags().StringVarP(&tanzuServer, "server", "s", "", "Address of the server to authenticate against.")
	c.PersistentFlags().StringVarP(&tanzuUsername, "username", "u", "", "Username to authenticate.")
	c.PersistentFlags().StringVarP(&tanzuPassword, "password", "p", "", "Password to use for authentication.")
	c.PersistentFlags().BoolVarP(&insecureSkipVerify, "insecure", "i", true, "Skip certificate verification (this is insecure).")
	c.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", fmt.Sprintf("%s/.kube/config", homedir), "Path to kubeconfig file.")

	// Setup sub-commands
	c.AddCommand(version.NewCmdVersion())
	c.AddCommand(login.NewCmdLogin())
	c.AddCommand(logout.NewCmdLogout())
	c.AddCommand(inspect.NewCmdInspect())
	c.AddCommand(list.NewCmdList())

	return c
}
