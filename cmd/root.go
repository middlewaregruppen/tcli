package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/middlewaregruppen/tcli/cmd/inspect"
	"github.com/middlewaregruppen/tcli/cmd/list"
	"github.com/middlewaregruppen/tcli/cmd/login"
	"github.com/middlewaregruppen/tcli/cmd/logout"
	"github.com/middlewaregruppen/tcli/cmd/use"
	"github.com/middlewaregruppen/tcli/cmd/version"
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
	debug              bool
	kubeconfig         string
	timeout            time.Duration
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
	export TCLI_INSECURE=true

	Use "tcli --help" for a list of global command-line options (applies to all commands).
	`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level := slog.LevelWarn
			if debug {
				level = slog.LevelDebug
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Check if kubeconfig exists, create if it doesn't
			if _, err := os.Stat(kubeconfig); errors.Is(err, os.ErrNotExist) {
				_, err = os.OpenFile(kubeconfig, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// Setup flags
	c.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "How long to wait for an operation before giving up")
	c.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging (HTTP traces written to stderr)")
	c.PersistentFlags().StringVarP(&tanzuServer, "server", "s", "", "Address of the server to authenticate against.")
	c.PersistentFlags().StringVarP(&tanzuUsername, "username", "u", "", "Username to authenticate.")
	c.PersistentFlags().StringVarP(&tanzuPassword, "password", "p", "", "Password to use for authentication.")
	c.PersistentFlags().BoolVarP(&insecureSkipVerify, "insecure", "i", false, "Skip certificate verification (this is insecure).")
	c.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", fmt.Sprintf("%s/.kube/config", homedir), "Path to kubeconfig file.")

	// Setup sub-commands
	c.AddCommand(version.NewCmdVersion())
	c.AddCommand(login.NewCmdLogin())
	c.AddCommand(logout.NewCmdLogout())
	c.AddCommand(inspect.NewCmdInspect())
	c.AddCommand(list.NewCmdList())
	c.AddCommand(use.NewCmdUse())

	return c
}
