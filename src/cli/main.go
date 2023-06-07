package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/levilutz/basiccoin/src/db"
	"github.com/levilutz/basiccoin/src/util"
)

// Define all commands available on this cli.
var commands = []Command{
	{
		Name:           "version",
		HelpText:       "Get the version of the cli.",
		RequiresClient: false,
		Handler: func(ctx HandlerContext) error {
			fmt.Println(util.Constants.Version)
			return nil
		},
	},
	{
		Name:           "connect",
		HelpText:       "Set up the remote address.",
		ArgsUsage:      "[address]",
		RequiredArgs:   1,
		RequiresClient: false,
		Handler: func(ctx HandlerContext) error {
			addr := ctx.Args[0]
			_, err := NewClient(addr)
			if err != nil {
				return fmt.Errorf("failed to connect to client: " + err.Error())
			}
			return NewConfig(addr).Save()
		},
	},
	{
		Name:           "import",
		HelpText:       "Import the given file into the current wallet.",
		ArgsUsage:      "[path]",
		RequiredArgs:   1,
		RequiresClient: false,
		Handler: func(ctx HandlerContext) error {
			newCfg := GetConfig(ctx.Args[0])
			newCfg.VerifyKeys()
			ctx.Config.AddKeys(newCfg.Keys...)
			return ctx.Config.Save()
		},
	},
	{
		Name:           "generate",
		HelpText:       "Generate a new address to receive basiccoin.",
		RequiresClient: false,
		Handler: func(ctx HandlerContext) error {
			priv, err := db.NewEcdsa()
			if err != nil {
				return err
			}
			kc := NewKeyConfig(priv)
			ctx.Config.AddKeys(kc)
			fmt.Printf("%x\n", kc.PublicKeyHash)
			return ctx.Config.Save()
		},
	},
	{
		Name:           "balance",
		HelpText:       "Get the total balance of all currently controlled addresses, or a given address.",
		ArgsUsage:      "(address)",
		RequiredArgs:   0,
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			if len(ctx.Args) > 0 {
				// Get balance of a given address
				pkh, err := db.StringToHash(ctx.Args[0])
				if err != nil {
					return err
				}
				balance, err := ctx.Client.GetBalance(pkh)
				if err != nil {
					return err
				}
				fmt.Printf("%x\t%d\n", pkh, balance)
			}
			return nil
		},
	},
	{
		Name:           "send",
		HelpText:       "Send coin to a given address.",
		ArgsUsage:      "[address] [amount]",
		RequiredArgs:   0,
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			return nil
		},
	},
	{
		Name:           "history",
		HelpText:       "Get the history of all currently controlled addresses, or a given address.",
		ArgsUsage:      "(address)",
		RequiredArgs:   0,
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			return nil
		},
	},
	{
		Name:           "get-config-path",
		HelpText:       "Print the path to our current config file.",
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			return nil
		},
	},
}

// Parse input and run commands as necessary.
func main() {
	Execute(commands)
}

// Read a line from stdin, given prompt.
func ReadInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return text[:len(text)-1], nil
}
