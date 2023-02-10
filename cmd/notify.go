package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// notifyCmd represents the notify command
var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "notify server, wait to implement",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wait to implement")
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.Flags().StringP("config", "c", "./config.yaml", "deployment services' config")
}
