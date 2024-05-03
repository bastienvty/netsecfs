package server

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server [mountDir]",
	Short: "Start the server",
	Long:  `Start the server that will mount handle the filesystem requests.`,
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
	ServerCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
}
