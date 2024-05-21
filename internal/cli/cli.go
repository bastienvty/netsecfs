package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/bastienvty/netsecfs/internal/db/object"
	"github.com/bastienvty/netsecfs/internal/share"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
)

const (
	input = "netsecfs> "
)

var (
	isMounted bool
)

func Initialize(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Please provide a mount point.")
		os.Exit(1)
	}
	addr, _ := cmd.Flags().GetString("meta")
	mp := args[0]

	m := meta.RegisterMeta(addr)
	format, err := m.Load()
	if err != nil {
		fmt.Println("Load fail: ", err)
		return
	}
	blob, err := object.CreateStorage(format.Storage)
	if err != nil {
		fmt.Println("CreateStorage fail: ", err)
		return
	}

	startConsole(m, blob, mp)
}

func startConsole(m meta.Meta, blob object.ObjectStorage, mp string) {
	scanner := bufio.NewScanner(os.Stdin)
	var server *fuse.Server
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
			if isMounted {
				fmt.Println("It may be still mounted. Please unmount it.")
			}
			return
		case "help":
			fmt.Println("Commands: mount, umount, share, exit")
		case "ls":
			fmt.Println("Not implemented but would list all users.")
		case "mount":
			server, err = mount(m, blob, mp)
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
			isMounted = true
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
			fmt.Println("Umount successfull.")
			isMounted = false
		case "share":
			share.Share(mp)
		}
	}
}
