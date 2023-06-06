package main

import (
	"fmt"
)

var commandHelpDispatch map[string]func() = map[string]func(){
	"setup": func() {
		fmt.Println("Set up the local wallet instance.")
		fmt.Println("usage: basiccoin-cli setup")
	},
	"import": func() {
		fmt.Println("Import the given file into the current wallet.")
		fmt.Println("usage: basiccoin-cli import [path]")
	},
	"balance": func() {
		fmt.Println("Get the total balance of all currently controlled addresses, or a given address.")
		fmt.Println("usage: basiccoin-cli balance (address)")
	},
	"send": func() {
		fmt.Println("Send coin to a given address.")
		fmt.Println("usage: basiccoin-cli send [address] [amount]")
	},
	"history": func() {
		fmt.Println("Get the history of all currently controlled addresses, or a given address.")
		fmt.Println("usage: basiccoin-cli history (address)")
	},
	"get-config-path": func() {
		fmt.Println("Print the path to our current config file.")
		fmt.Println("usage: basiccoin-cli config-dir")
	},
}

func PrintGeneralHelp() {
	fmt.Println("Query a basiccoin node and manage a wallet.")
	fmt.Println("usage: basiccoin-cli [command] ...")
	fmt.Println("available commands:")
	for cmd := range commandHelpDispatch {
		fmt.Printf("\t%s\n", cmd)
	}
}

func PrintCommandHelp(command string) {
	helpFunc, ok := commandHelpDispatch[command]
	if !ok {
		fmt.Println(yellowStr("unknown command: " + command))
	} else {
		helpFunc()
	}
}

func greenStr(str string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", str)
}

func yellowStr(str string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", str)
}

func redStr(str string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", str)
}
