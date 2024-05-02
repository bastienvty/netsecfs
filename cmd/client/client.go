/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start the client",
	Long:  `Start the client that will handle the filesystem requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
