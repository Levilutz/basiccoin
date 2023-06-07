package main

import (
	"fmt"
	"os"
)

// Context to be passed to command handler functions.
type HandlerContext struct {
	Args   []string
	Config *Config
	Client *Client
}

// A single command with its help, requirements, and handler function.
type Command struct {
	Name           string
	HelpText       string
	ArgsUsage      string
	RequiredArgs   int
	RequiresClient bool
	Handler        func(ctx HandlerContext) error
}

// Create the general usage text string.
func (cmd Command) UsageText() string {
	return fmt.Sprintf("Usage: basiccoin-cli %s %s", cmd.Name, cmd.ArgsUsage)
}

func Execute(commands []Command) {
	// Convert commands to map
	cmdMap := make(map[string]Command, len(commands))
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	// Ensure config exists, then load it
	EnsureConfig()
	cfg := GetConfig(getConfigPath())
	cfg.VerifyKeys()

	// Try to make client from the config
	var client *Client = nil
	if cfg.NodeAddr != "" {
		tryClient, err := NewClient(cfg)
		if err != nil {
			fmt.Println(yellowStr("failed to connect to configured node: " + err.Error()))
		} else {
			client = tryClient
		}
	}

	// Get cli args
	if len(os.Args) < 2 {
		fmt.Println(yellowStr("must provide command"))
		return
	}
	command := os.Args[1]
	cmdArgs := os.Args[2:]

	// Show general help message if wanted
	if command == "help" {
		printGeneralHelp()
		return
	}

	// Get the command and verify it exists
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

	// Verify client connected if command requires it
	if client == nil && cmd.RequiresClient {
		fmt.Println(yellowStr("command requires valid node connection, run 'basiccoin-cli connect' to set up"))
		return
	}

	// Run the command
	err := cmd.Handler(HandlerContext{
		Args:   cmdArgs,
		Config: cfg,
		Client: client,
	})
	if err != nil {
		fmt.Println(redStr(err.Error()))
	} else {
		fmt.Println(greenStr("success"))
	}
}

// Print general help on this CLI.
func printGeneralHelp() {
	fmt.Println("Query a basiccoin node and manage a wallet.")
	fmt.Println("Usage: basiccoin-cli [command] ...")
	fmt.Println("Available commands:")
	for _, cmd := range commands {
		fmt.Printf(" - %s\n", cmd.Name)
	}
	fmt.Println("For more help, run 'basiccoin-cli [command] help'")
}

// Turn a string green.
func greenStr(str string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", str)
}

// Turn a string yellow.
func yellowStr(str string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", str)
}

// Turn a string red.
func redStr(str string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", str)
}
