package server

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server [mountDir]",
	Short: "Start the server",
	Long:  `Start the server that will mount handle the filesystem requests.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Error while parsing the debug flag.")
			return
		}
		mount(args[0], debug)
	},
}

func init() {
	ServerCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
}
