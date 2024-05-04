package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount the filesystem",
	Long:  `Mount the filesystem to the specified directory.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		/*debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Error while parsing the debug flag.")
			return
		}*/
		server := mount(args[0])
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			server.Serve()
			wg.Done()
		}()

		fmt.Printf("Unmount to terminate.\n")
		fmt.Printf("\n")

		wg.Wait()
	},
}

func init() {
	mountCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
}
