package main

import (
	"fmt"
	"os"
)

// Context to be passed to command handler functions.
type HandlerContext struct {
	Args   []string
	Config *Config
}

// A single command with its help, requirements, and handlerr function.
type Command struct {
	Name         string
	HelpText     string
	ArgsUsage    string
	RequiredArgs int
	Handler      func(ctx *HandlerContext) error
}

// Create the general usage text string.
func (cmd Command) UsageText() string {
	return fmt.Sprintf("Usage: bcwallet %s %s", cmd.Name, cmd.ArgsUsage)
}

func Execute(commands []Command) {
	// Convert commands to map
	cmdMap := make(map[string]Command, len(commands))
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	// Get cli args
	if len(os.Args) < 2 {
		fmt.Println(yellowStr("must provide command"))
		return
	}
	var dev bool
	var command string
	var cmdArgs []string
	if os.Args[1] == "dev" {
		dev = true
		command = os.Args[2]
		cmdArgs = os.Args[3:]
	} else {
		dev = false
		command = os.Args[1]
		cmdArgs = os.Args[2:]
	}

	// Ensure config exists, then load it
	EnsureConfig(dev)
	cfg := GetConfig(getConfigPath(dev))
	cfg.VerifyKeys()

	// Show general help message if wanted
	if command == "help" {
		printGeneralHelp(commands)
		return
	}

	// Get the command provided it exists
	cmd, ok := cmdMap[command]
	if !ok {
		fmt.Println(yellowStr(fmt.Sprintf("command '%s' not found", command)))
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

	// Run the command
	err := cmd.Handler(&HandlerContext{
		Args:   cmdArgs,
		Config: cfg,
	})
	if err != nil {
		fmt.Println(redStr(err.Error()))
	} else {
		fmt.Println(greenStr("success"))
	}
}

func printGeneralHelp(commands []Command) {
	fmt.Println("Manage a basiccoin wallet.")
	fmt.Println("Usage: bcwallet [command] ...")
	fmt.Println("Available commands")
	for _, cmd := range commands {
		fmt.Printf(" - %s\n", cmd.Name)
	}
	fmt.Println("For more help, run 'bcwallet [command] help'")
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
