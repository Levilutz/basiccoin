package main

import (
	"fmt"
	"os"
)

type Command struct {
	Name           string
	HelpText       string
	ArgsUsage      string
	RequiredArgs   int
	RequiresConfig bool
	Handler        func(args []string, cfg *Config) error
}

func (cmd Command) UsageText() string {
	return fmt.Sprintf("Usage: basiccoin-cli %s %s", cmd.Name, cmd.ArgsUsage)
}

var commands = []Command{
	{
		Name:           "setup",
		HelpText:       "Set up the local wallet instance.",
		ArgsUsage:      "",
		RequiredArgs:   0,
		RequiresConfig: false,
		Handler: func(args []string, cfg *Config) error {
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

func main() {
	// Convert commands to map
	cmdMap := make(map[string]Command)
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	// Load config, if it exists
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

	cmd, ok := cmdMap[command]
	if !ok {
		fmt.Println(yellowStr("command not found"))
		return
	}

	// Show command help message if wanted
	if len(cmdArgs) > 0 && cmdArgs[0] == "help" {
		fmt.Println(cmd.HelpText)
		fmt.Println(cmd.UsageText())
		return
	}

	// Verify sufficient arguments
	if len(cmdArgs) < cmd.RequiredArgs {
		fmt.Println(yellowStr("insufficient arguments"))
		fmt.Println(cmd.UsageText())
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
	fmt.Println("Usage: basiccoin-cli [command] ...")
	fmt.Println("Available commands:")
	for _, cmd := range commands {
		fmt.Printf(" - %s\n", cmd.Name)
	}
	fmt.Println("For more help, run 'basiccoin-cli [command] help'")
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
