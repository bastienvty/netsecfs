package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bastienvty/netsecfs/internal/crypto"
	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/bastienvty/netsecfs/internal/db/object"
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

	startConsole(m, blob, mp)
}

func startConsole(m meta.Meta, blob object.ObjectStorage, mp string) {
	scanner := bufio.NewScanner(os.Stdin)
	var server *fuse.Server
	var err error
	var user User
	for {
		fmt.Print(input)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "exit":
			if isMounted {
				fmt.Println("It may be still mounted. Please unmount it.")
			}
			return
		case "help":
			fmt.Println("Commands: signup, login, logout, passwd, mount, umount, share, unshare and exit")
		case "signup":
			if isLogged {
				fmt.Println("User already logged in.")
				continue
			}
			if len(fields) != 3 {
				fmt.Println("Usage: signup <username> <password>")
				continue
			}
			user = User{
				username: fields[1],
				password: fields[2],
				m:        m,
				enc:      &crypto.CryptoHelper{},
			}
			// startTime := time.Now()
			create := user.createUser()
			if !create {
				fmt.Println("User creation failed. Please try again.")
				continue
			}
			// duration := time.Since(startTime)
			// fmt.Printf("The signup took %s to complete.\n", duration)
			fmt.Printf("User %s created.\n", user.username)
			isLogged = true
		case "login":
			if isLogged {
				fmt.Println("User already logged in.")
				continue
			}
			if len(fields) != 3 {
				fmt.Println("Usage: login <username> <password>")
				continue
			}
			user = User{
				username: fields[1],
				password: fields[2],
				m:        m,
				enc:      &crypto.CryptoHelper{},
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
		case "mount":
			if isMounted {
				fmt.Println("Already mounted.")
				continue
			}
			if !isLogged {
				fmt.Println("User not logged in.")
				continue
			}
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
			if !isMounted {
				fmt.Println("Mount before sharing.")
				continue
			}
			if !isLogged {
				fmt.Println("User not logged in.")
				continue
			}
			if len(fields) != 3 {
				fmt.Println("Usage: share <folder_path> <user>")
				continue
			}
			dir := mp + "/" + fields[1]
			shared := user.shareDir(dir, fields[2])
			if !shared {
				fmt.Println("Share failed. Please try again.")
				continue
			}
			fmt.Println("Share successfull.")
		case "unshare":
			if !isMounted {
				fmt.Println("Mount before unsharing.")
				continue
			}
			if !isLogged {
				fmt.Println("User not logged in.")
				continue
			}
			if len(fields) != 3 {
				fmt.Println("Usage: unshare <folder_path> <user>")
				continue
			}
			dir := mp + "/" + fields[1]
			unshared := user.unshareDir(dir, fields[2])
			if !unshared {
				fmt.Println("Unshare failed. Please try again.")
				continue
			}
			fmt.Println("Unshare successfull.")
		case "logout":
			if !isLogged {
				fmt.Println("User not logged in.")
				continue
			}
			if isMounted {
				fmt.Println("Unmount before logging out.")
				continue
			}
			fmt.Printf("User %s logged out.\n", user.username)
			isLogged = false
			user = User{}
		}
	}
}
