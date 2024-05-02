package server

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server that will handle the filesystem requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("server called")
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")
	// serverCmd.PersistentFlags().String("bar", "", "A help for bar")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
