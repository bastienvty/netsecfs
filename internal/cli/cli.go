package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/bastienvty/netsecfs/internal/db/object"
	"github.com/bastienvty/netsecfs/internal/fs"
	"github.com/bastienvty/netsecfs/internal/share"
	gofs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
)

const (
	input = "netsecfs> "
)

func StartConsole(cmd *cobra.Command, args []string) {
	scanner := bufio.NewScanner(os.Stdin)
	var server *fuse.Server
	var m meta.Meta
	var blob object.ObjectStorage
	var err error
	for {
		fmt.Print(input)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		switch line {
		case "exit":
			return
		case "help":
			fmt.Println("Commands: mount, umount, share, exit")
		case "mount":
			m, blob, server, err = mount(cmd, args)
			if err != nil || server == nil {
				fmt.Println("Mount fail: ", err)
				return
			}
			if m != nil {
				defer m.Shutdown()
			}
			if blob != nil {
				defer object.Shutdown(blob)
			}
		case "umount":
			if server == nil {
				fmt.Println("Server is nil.")
				continue
			}
			err = server.Unmount()
			if err != nil {
				fmt.Println("Unmount fail: ", err)
				continue
			}
		case "share":
			share.Share(args[0])
		}
		fmt.Println("You entered:", line)
	}
}

func mount(cmd *cobra.Command, args []string) (meta.Meta, object.ObjectStorage, *fuse.Server, error) {
	addr, _ := cmd.Flags().GetString("meta")
	mp := args[0]

	m := meta.RegisterMeta(addr)
	format, err := m.Load()
	if err != nil {
		fmt.Println("Load fail: ", err)
		return nil, nil, nil, err
	}

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
		UID: uint32(os.Getuid()),
		GID: uint32(os.Getgid()),
	}
	fuseOpts.MountOptions = fuse.MountOptions{
		Options: []string{"rw", "default_permissions"},
		Debug:   false,
		Name:    "netsecfs",
	}
	fuseOpts.MountOptions.Options = append(fuseOpts.MountOptions.Options, "noapplexattr", "noappledouble") // macOS

	blob, err := object.CreateStorage(format.Storage)
	if err != nil {
		fmt.Println("CreateStorage fail: ", err)
		return nil, nil, nil, err
	}

	syscall.Umask(0000)
	root := fs.NewRootNode(m, blob)
	server, err := gofs.Mount(mp, root, fuseOpts)
	if err != nil {
		fmt.Println("Mount fail: ", err)
		return nil, nil, nil, err
	}

	// server.Wait()
	// fmt.Println("Server exited.")
	fmt.Println("Unmount to stop the server.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.Unmount()
	}()
	return m, blob, server, nil
}
