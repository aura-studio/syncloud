/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/aura-studio/syncloud/pusher"

	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var c pusher.Config

		if remotes, err := cmd.Flags().GetStringSlice("remote"); err != nil {
			log.Panic(err)
		} else if len(remotes) > 0 {
			c.Remotes = remotes
		}

		if locals, err := cmd.Flags().GetStringSlice("local"); err != nil {
			log.Panic(err)
		} else if len(locals) > 0 {
			c.Locals = locals
		}

		pusher.New(pusher.NewTaskList(c)).Push()
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringSliceP("local", "l", nil, "local file path")
	pushCmd.Flags().StringSliceP("remote", "r", nil, "remote file path")
}
