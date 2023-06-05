package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(redStr("must provide command"))
		return
	}
	command := os.Args[1]
	cmdArgs := os.Args[2:]
	helpWanted := len(os.Args) >= 1 && cmdArgs[0] == "help"

	if command == "help" {
		fmt.Println("Query a basiccoin node and manage a wallet.")
		fmt.Println("usage: basiccoin-cli [command] ...")
		fmt.Println("available commands:")
		for _, cmd := range []string{"login", "balance", "send", "history", "export", "import"} {
			fmt.Printf("\t%s\n", cmd)
		}

	} else if command == "setup" {
		if len(cmdArgs) < 1 || helpWanted {
			fmt.Println("Set up the local wallet instance.")
			fmt.Println("usage: basiccoin-cli setup")
			return
		}
		fmt.Println(greenStr("success"))

	} else if command == "balance" {
		if helpWanted {
			fmt.Println("Get the total balance of all currently controlled addresses, or a given address.")
			fmt.Println("usage: basiccoin-cli balance (address)")
			return
		}

	} else if command == "send" {
		if len(cmdArgs) < 2 || helpWanted {
			fmt.Println("Send coin to a given address.")
			fmt.Println("usage: basiccoin-cli send [address] [amount]")
			return
		}
		fmt.Println(greenStr("success"))

	} else if command == "history" {
		if helpWanted {
			fmt.Println("Get the history of all currently controlled addresses, or a given address.")
			fmt.Println("usage: basiccoin-cli history (address)")
			return
		}

	} else if command == "export" {
		if len(cmdArgs) < 1 || helpWanted {
			fmt.Println("Export the current wallet to a given file.")
			fmt.Println("usage: basiccoin-cli export [path]")
			return
		}
		fmt.Println(greenStr("success"))

	} else if command == "import" {
		if len(cmdArgs) < 1 || helpWanted {
			fmt.Println("Import the given file into the current wallet.")
			fmt.Println("usage: basiccoin-cli import [path]")
		}
		fmt.Println(greenStr("success"))

	} else {
		fmt.Println(yellowStr("command not found:"), command)
	}
}
