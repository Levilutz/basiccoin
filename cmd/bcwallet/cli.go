package main

import (
	"fmt"
	"os"
)

// Context to be passed to command handler functions.
type HandlerContext struct {
	Args []string
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
	command := os.Args[1]
	cmdArgs := os.Args[2:]

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
		Args: cmdArgs,
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
