package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bastienvty/netsecfs/internal/crypto"
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
	isLogged  bool
)

func Initialize(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Please provide a mount point.")
		os.Exit(1)
	}
	addr, _ := cmd.Flags().GetString("meta")
	username, _ := cmd.Flags().GetString("username")
	pwd, _ := cmd.Flags().GetString("password")
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
	if m != nil {
		defer m.Shutdown()
	}
	if blob != nil {
		defer object.Shutdown(blob)
	}

	user := User{
		username: username,
		password: pwd,
		m:        m,
		enc:      crypto.CryptoHelper{},
	}

	startConsole(user, blob, mp)
}

func startConsole(user User, blob object.ObjectStorage, mp string) {
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
		fields := strings.Fields(line)
		switch fields[0] {
		case "exit":
			if isMounted {
				fmt.Println("It may be still mounted. Please unmount it.")
			}
			return
		case "help":
			fmt.Println("Commands: mount, umount, share, exit")
		case "signup":
			create := user.createUser()
			if !create {
				fmt.Println("User creation failed. Please try again.")
				continue
			}
			fmt.Printf("User %s created.\n", user.username)
			isLogged = true
		case "login":
			if isLogged {
				fmt.Println("User already logged in.")
				continue
			}
			verify := user.verifyUser()
			if !verify {
				fmt.Println("User verification failed. Please try again.")
				continue
			}
			isLogged = true
			fmt.Printf("User %s logged in.\n", user.username)
		case "passwd":
			if !isLogged {
				fmt.Println("User not logged in.")
				continue
			}
			if isMounted {
				fmt.Println("Unmount before changing password.")
				continue
			}
			if len(fields) != 2 {
				fmt.Println("Usage: passwd <new_password>")
				continue
			}
			changed := user.changePassword(fields[1])
			if !changed {
				fmt.Println("Password change failed. Please try again.")
				continue
			}
			fmt.Println("Password changed successfully.")
		case "ls":
			fmt.Println("Not implemented but would list all users.")
		case "mount":
			server, err = mount(user, blob, mp)
			if err != nil || server == nil {
				fmt.Println("Mount fail: ", err)
				return
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
		case "unshare":
			// share.Unshare(mp)
		}
	}
}
