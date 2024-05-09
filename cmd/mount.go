package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/bastienvty/netsecfs/internal/fs"
	gofs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
)

var (
	user string
	pwd  string
)

var mountCmd = &cobra.Command{
	Use:       "mount [flags] MOUNTPOINT",
	Short:     "Mount the filesystem",
	Long:      `Mount the filesystem to the specified directory.`,
	ValidArgs: []string{"username", "password"},
	Args:      cobra.ExactArgs(1),
	Example:   "netsecfs mount --storage /path/to/storage --meta /path/to/meta.db --username toto --password titi /tmp/nsfs",
	Run:       mount,
}

func mount(cmd *cobra.Command, args []string) {
	var fuseOpts *gofs.Options
	sec := time.Second
	fuseOpts = &gofs.Options{
		// These options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		NegativeTimeout: &sec,
		AttrTimeout:     &sec,
		EntryTimeout:    &sec,
		UID:             uint32(os.Getuid()),
		GID:             uint32(os.Getgid()),
	}
	fuseOpts.MountOptions = fuse.MountOptions{
		Options: []string{"rw", "default_permissions"},
		//Debug:   true,
	}
	//fuseOpts.MountOptions.Options = append(fuseOpts.MountOptions.Options, "rw")
	server, err := gofs.Mount(args[0], &fs.MyNode{}, fuseOpts)
	if err != nil {
		fmt.Println("Mount fail: ", err)
		return
	}
	fmt.Println("Unmount to stop the server.")
	server.Wait()
	//go server.Serve()
}

func init() {
	/*initCmd.Flags().StringP("storage", "s", "", "Path to the storage database.")
	initCmd.Flags().StringP("meta", "m", "", "Path to the meta database.")
	initCmd.MarkFlagRequired("storage")
	initCmd.MarkFlagRequired("meta")*/

	mountCmd.Flags().StringVarP(&user, "username", "u", "", "Username")
	mountCmd.Flags().StringVarP(&pwd, "password", "p", "", "Password")
	mountCmd.MarkFlagRequired("username")
	mountCmd.MarkFlagsRequiredTogether("username", "password")
}
