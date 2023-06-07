package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/levilutz/basiccoin/src/util"
)

// Define all commands available on this cli.
var commands = []Command{
	{
		Name:           "version",
		HelpText:       "Get the version of the cli.",
		ArgsUsage:      "",
		RequiredArgs:   0,
		RequiresConfig: false,
		Handler: func(args []string, cfg *Config) error {
			fmt.Println(util.Constants.Version)
			return nil
		},
	},
	{
		Name:           "setup",
		HelpText:       "Set up the local wallet instance.",
		ArgsUsage:      "",
		RequiredArgs:   0,
		RequiresConfig: false,
		Handler: func(args []string, cfg *Config) error {
			addr, err := readInput("Node address: ")
			if err != nil {
				return err
			}
			fmt.Printf("<%s>\n", addr)
			return nil
		},
	},
	{
		Name:           "import",
		HelpText:       "Import the given file into the current wallet.",
		ArgsUsage:      "[path]",
		RequiredArgs:   0,
		RequiresConfig: false,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	{
		Name:           "generate",
		HelpText:       "Generate a new address to receive basiccoin.",
		ArgsUsage:      "",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	{
		Name:           "balance",
		HelpText:       "Get the total balance of all currently controlled addresses, or a given address.",
		ArgsUsage:      "(address)",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	{
		Name:           "send",
		HelpText:       "Send coin to a given address.",
		ArgsUsage:      "[address] [amount]",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	{
		Name:           "history",
		HelpText:       "Get the history of all currently controlled addresses, or a given address.",
		ArgsUsage:      "(address)",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	{
		Name:           "get-config-path",
		HelpText:       "Print the path to our current config file.",
		ArgsUsage:      "",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
}

// Parse input and run commands as necessary.
func main() {
	Execute(commands)
}

// Read a line from stdin, given prompt.
func readInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return text[:len(text)-1], nil
}
