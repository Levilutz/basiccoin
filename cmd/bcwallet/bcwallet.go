package main

import (
	"fmt"

	"github.com/levilutz/basiccoin/pkg/core"
)

// Define commands available on this cli.
var commands = []Command{
	{
		Name:           "version",
		HelpText:       "Get the version of the cli.",
		RequiresClient: false,
		Handler: func(ctx *HandlerContext) error {
			fmt.Println(ctx.Config.Version())
			return nil
		},
	},
	{
		Name:           "get-config-path",
		HelpText:       "Get the path to our current config file.",
		RequiresClient: false,
		Handler: func(ctx *HandlerContext) error {
			fmt.Println(getConfigPath(ctx.Config.Dev))
			return nil
		},
	},
	{
		Name:           "import",
		HelpText:       "Import the given file into the current wallet.",
		ArgsUsage:      "[path]",
		RequiredArgs:   1,
		RequiresClient: false,
		Handler: func(ctx *HandlerContext) error {
			newCfg := GetConfig(ctx.Args[0])
			newCfg.VerifyKeys()
			ctx.Config.AddKeys(newCfg.Keys...)
			return ctx.Config.Save()
		},
	},
	{
		Name:           "connect",
		HelpText:       "Set up the remote addresss.",
		ArgsUsage:      "[address]",
		RequiredArgs:   1,
		RequiresClient: false,
		Handler: func(ctx *HandlerContext) error {
			addr := ctx.Args[0]
			_, err := NewClient(&Config{
				Dev:      ctx.Config.Dev,
				NodeAddr: addr,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to client: " + err.Error())
			}
			ctx.Config.NodeAddr = addr
			return ctx.Config.Save()
		},
	},
	{
		Name:           "generate",
		HelpText:       "Generate a new address to receive coin.",
		RequiresClient: false,
		Handler: func(ctx *HandlerContext) error {
			priv, err := core.NewEcdsa()
			if err != nil {
				return err
			}
			kc := NewKeyConfig(priv)
			ctx.Config.AddKeys(kc)
			fmt.Println(kc.PublicKeyHash)
			return ctx.Config.Save()
		},
	},
}

func main() {
	Execute(commands)
}
