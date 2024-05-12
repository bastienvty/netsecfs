package cmd

import (
	"fmt"
	"syscall"
	"time"

	"github.com/bastienvty/netsecfs/internal/db/meta"
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
	Example:   "netsecfs mount --meta /path/to/meta.db --username toto --password titi /tmp/nsfs",
	Run:       mount,
}

func mount(cmd *cobra.Command, args []string) {
	addr, _ := cmd.Flags().GetString("meta")
	mp := args[0]

	m := meta.RegisterMeta(addr)
	/*format, err := m.Load()
	if err != nil {
		fmt.Println("Load fail: ", err)
		return
	}*/

	var fuseOpts *gofs.Options
	sec := time.Second
	fuseOpts = &gofs.Options{
		// These options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		NegativeTimeout: &sec,
		AttrTimeout:     &sec,
		EntryTimeout:    &sec,
		RootStableAttr: &gofs.StableAttr{
			Ino: uint64(meta.RootInode),
		},
		//UID:             uint32(os.Getuid()),
		//GID:             uint32(os.Getgid()),
	}
	fuseOpts.MountOptions = fuse.MountOptions{
		Options: []string{"rw", "default_permissions"},
		Debug:   true,
		Name:    "netsecfs",
	}
	fuseOpts.MountOptions.Options = append(fuseOpts.MountOptions.Options, "noapplexattr", "noappledouble") // macOS

	/*blob, err := object.CreateStorage(format.Storage)
	if err != nil {
		fmt.Println("CreateStorage fail: ", err)
		return
	}

	if m != nil {
		if err = m.Shutdown(); err != nil {
			logger.Errorf("[pid=%d] meta shutdown: %s", os.Getpid(), err)
		}
	}
	if blob != nil {
		object.Shutdown(blob)
	}*/

	//fuseOpts.MountOptions.Options = append(fuseOpts.MountOptions.Options, "rw")
	syscall.Umask(0000)
	root := fs.NewNode(m)
	server, err := gofs.Mount(mp, root, fuseOpts)
	if err != nil {
		fmt.Println("Mount fail: ", err)
		return
	}
	fmt.Println("Unmount to stop the server.")
	/*c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.Unmount()
	}()*/

	server.Wait()
	fmt.Println("Server exited.")
	//go server.Serve()
}

func init() {
	mountCmd.Flags().StringP("meta", "m", "", "Path to the meta database.")
	mountCmd.MarkFlagRequired("meta")

	mountCmd.Flags().StringVarP(&user, "username", "u", "", "Username")
	mountCmd.Flags().StringVarP(&pwd, "password", "p", "", "Password")
	mountCmd.MarkFlagRequired("username")
	mountCmd.MarkFlagsRequiredTogether("username", "password")
}
