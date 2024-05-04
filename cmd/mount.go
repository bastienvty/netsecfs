package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	user string
	pwd  string
)

var mountCmd = &cobra.Command{
	Use:       "mount [flags] META-URL MOUNTPOINT",
	Short:     "Mount the filesystem",
	Long:      `Mount the filesystem to the specified directory.`,
	ValidArgs: []string{"username", "password"},
	Args:      cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Error while parsing the debug flag.")
			return
		}
		server := mount(args[1], debug)
		fmt.Println("Unmount to stop the server.")
		server.Wait()
	},
}

func init() {
	mountCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	mountCmd.Flags().StringVarP(&user, "username", "u", "", "Username")
	mountCmd.Flags().StringVarP(&pwd, "password", "p", "", "Password")
	mountCmd.MarkFlagRequired("username")
	mountCmd.MarkFlagsRequiredTogether("username", "password")
}
