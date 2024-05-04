package cli

import (
	"bufio"
	"fmt"
	"os"
)

func StartConsole() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		if line == "exit" {
			break
		}
		fmt.Println("You entered:", line)
	}
	fmt.Println("Exiting console...")
}
