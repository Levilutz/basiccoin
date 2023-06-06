package main

import (
	"fmt"
	"os"
)

func main() {
	cfg := GetConfig()

	// Get cli args
	if len(os.Args) < 2 {
		fmt.Println(yellowStr("must provide command"))
		return
	}
	command := os.Args[1]
	cmdArgs := os.Args[2:]

	// Show help message if wanted
	if command == "help" {
		PrintGeneralHelp()
		return
	} else if len(cmdArgs) > 0 && cmdArgs[0] == "help" {
		PrintCommandHelp(command)
		return
	}

	// Commands that don't require config
	if command == "setup" {
		fmt.Println(greenStr("success"))
		return
	}

	if command == "import" {
		if len(cmdArgs) < 1 {
			fmt.Println(yellowStr("insufficient arguments"))
			return
		}
		fmt.Println(greenStr("success"))
		return
	}

	if cfg == nil {
		fmt.Println(yellowStr("command requires setup, run 'basiccoin-cli setup' first"))
		return
	}

	// Commands that require config
	if command == "balance" {
		return

	} else if command == "send" {
		if len(cmdArgs) < 2 {
			fmt.Println(yellowStr("insufficient arguments"))
			return
		}
		fmt.Println(greenStr("success"))
		return

	} else if command == "history" {
		return

	} else if command == "config-dir" {
		return

	} else {
		fmt.Println(redStr("command not found:"), command)
	}
}
