/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/bastienvty/netsecfs/internal/cli"
	"github.com/spf13/cobra"
)

var (
	user string
	pwd  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "netsecfs",
	Short: "Mount a FUSE filesystem that encrypts files on the fly.",
	Long: `This filesystem allows you to mount a FUSE filesystem
that encrypts and decrypts files based on a password.

Be sure you have initialized the filesystem with the
init command before mounting it.`,
	ValidArgs: []string{"meta"},
	Args:      cobra.ExactArgs(1),
	Example:   "netsecfs --meta /path/to/meta.db /tmp/nsfs",
	Run: func(cmd *cobra.Command, args []string) {
		cli.Initialize(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print the version number of netsecfs")

	rootCmd.AddCommand(initCmd)

	rootCmd.Flags().StringP("meta", "m", "", "Path to the meta database.")
	rootCmd.MarkFlagRequired("meta")
}
