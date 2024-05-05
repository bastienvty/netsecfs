package cmd

import (
	"fmt"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
)

var (
	user string
	pwd  string
)

type RootNode struct {
	fs.Inode
}

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
	var fuseOpts *fs.Options
	sec := time.Second
	fuseOpts = &fs.Options{
		// These options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		NegativeTimeout: &sec,
		AttrTimeout:     &sec,
		EntryTimeout:    &sec,
	}
	fuseOpts.MountOptions = fuse.MountOptions{
		AllowOther: true,
		Debug:      true,
	}
	server, err := fs.Mount(args[0], &RootNode{}, fuseOpts)
	if err != nil {
		fmt.Println("Mount fail: ", err)
		return
	}
	fmt.Println("Unmount to stop the server.")
	server.Wait()
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