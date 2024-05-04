package cmd

import (
	"github.com/bastienvty/netsecfs/internal/cli"
	"github.com/spf13/cobra"
)

// initCmd represents the client command
var initCmd = &cobra.Command{
	Use:   "client",
	Short: "Start the client",
	Long:  `Start the client that will handle the filesystem requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.StartConsole()
	},
}

func init() {
	/*rootCmd.Flags().StringVarP(&u, "username", "u", "", "Username (required if password is set)")
	rootCmd.Flags().StringVarP(&pw, "password", "p", "", "Password (required if username is set)")
	rootCmd.MarkFlagsRequiredTogether("username", "password")*/
}
