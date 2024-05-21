package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/bastienvty/netsecfs/internal/db/object"
	"github.com/bastienvty/netsecfs/internal/fs"
	gofs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func mount(user User, blob object.ObjectStorage, mp string) (*fuse.Server, error) {
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
	// fuseOpts.MountOptions.Options = append(fuseOpts.MountOptions.Options, "noapplexattr", "noappledouble") // macOS (optional)

	syscall.Umask(0000)
	root := fs.NewRootNode(user.m, blob, user.masterKey, user.rootKey)
	server, err := gofs.Mount(mp, root, fuseOpts)
	if err != nil {
		fmt.Println("Mount fail: ", err)
		return nil, err
	}

	fmt.Println("Unmount to stop the server.")
	// server.Wait()
	// fmt.Println("Server exited.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.Unmount()
	}()
	return server, nil
}
