package share

import (
	"fmt"
	"os"
	"syscall"
)

func Share(dir string) {
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		fmt.Println("Not a syscall.Stat_t")
		return
	}

	fmt.Println("Inode Number:", stat.Ino)
	fmt.Println("Size:", stat.Size)
	fmt.Println("Number of Links:", stat.Nlink)
	fmt.Println("Permissions:", info.Mode())
	fmt.Println("Last Modified:", info.ModTime())
}
