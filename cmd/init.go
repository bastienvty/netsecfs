package cmd

import (
	"github.com/bastienvty/netsecfs/internal/cli"
	"github.com/spf13/cobra"
)

// initCmd represents the client command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the filesystem.",
	Long: `Initialize the filesystem by creating all necessary 
databases.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.StartConsole()
	},
}

func init() {

}
