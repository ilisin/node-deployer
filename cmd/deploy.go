package cmd

import (
	"os"

	"github.com/ilisin/node-deployer/deployer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFileDeploy string
)

// deployCmd represents the deployment command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy a service",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.WithError(err).Fatal("parse config file fail")
			os.Exit(1)
		}
		svr, err := deployer.NewServer(cfg)
		if err != nil {
			logrus.WithError(err).Fatal("init server fail")
			os.Exit(1)
		}
		svr.Run()
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("config", "c", "./config.yml", "deployment services' config")
}
