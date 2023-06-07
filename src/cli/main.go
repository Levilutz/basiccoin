package main

import (
	"fmt"
	"os"
)

type Command struct {
	HelpText       string
	UsageText      string
	RequiredArgs   int
	RequiresConfig bool
	Handler        func(args []string, cfg *Config) error
}

var commands map[string]Command = map[string]Command{
	"setup": {
		HelpText:       "Set up the local wallet instance.",
		UsageText:      "usage: basiccoin-cli setup",
		RequiredArgs:   0,
		RequiresConfig: false,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	"import": {
		HelpText:       "Import the given file into the current wallet.",
		UsageText:      "usage: basiccoin-cli import [path]",
		RequiredArgs:   0,
		RequiresConfig: false,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	"balance": {
		HelpText:       "Get the total balance of all currently controlled addresses, or a given address.",
		UsageText:      "usage: basiccoin-cli balance (address)",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	"send": {
		HelpText:       "Send coin to a given address.",
		UsageText:      "usage: basiccoin-cli send [address] [amount]",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	"history": {
		HelpText:       "Get the history of all currently controlled addresses, or a given address.",
		UsageText:      "usage: basiccoin-cli history (address)",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
	"get-config-path": {
		HelpText:       "Print the path to our current config file.",
		UsageText:      "usage: basiccoin-cli config-dir",
		RequiredArgs:   0,
		RequiresConfig: true,
		Handler: func(args []string, cfg *Config) error {
			return nil
		},
	},
}

func main() {
	cfg := GetConfig()

	// Get cli args
	if len(os.Args) < 2 {
		fmt.Println(yellowStr("must provide command"))
		return
	}
	command := os.Args[1]
	cmdArgs := os.Args[2:]

	// Show general help message if wanted
	if command == "help" {
		PrintGeneralHelp()
		return
	}

	cmd, ok := commands[command]
	if !ok {
		fmt.Println(yellowStr("command not found"))
		return
	}

	// Show command help message if wanted
	if len(cmdArgs) > 0 && cmdArgs[0] == "help" {
		fmt.Println(cmd.HelpText)
		fmt.Println(cmd.UsageText)
		return
	}

	// Verify sufficient arguments
	if len(cmdArgs) < cmd.RequiredArgs {
		fmt.Println(yellowStr("insufficient arguments"))
		fmt.Println(cmd.UsageText)
		return
	}

	// Verify configured if required
	if cfg == nil && cmd.RequiresConfig {
		fmt.Println(yellowStr("command requires setup, run 'basiccoin-cli setup' first"))
		return
	}

	// Run the command
	err := cmd.Handler(cmdArgs, cfg)
	if err != nil {
		fmt.Println(redStr(err.Error()))
	} else {
		fmt.Println(greenStr("success"))
	}
}

func PrintGeneralHelp() {
	fmt.Println("Query a basiccoin node and manage a wallet.")
	fmt.Println("usage: basiccoin-cli [command] ...")
	fmt.Println("available commands:")
	for cmd := range commands {
		fmt.Printf("\t%s\n", cmd)
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
