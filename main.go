package main

import (
	"os"

	"github.com/middlewaregruppen/tcli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.AutomaticEnv()
	if err := cmd.NewDefaultCommand().Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}
