package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/levilutz/basiccoin/internal/rest/client"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/set"
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
			_, err := client.NewWalletClient(ctx.Args[0], ctx.Config.Version())
			if err != nil {
				return fmt.Errorf("failed to connect to client: " + err.Error())
			}
			ctx.Config.NodeAddr = ctx.Args[0]
			return ctx.Config.Save()
		},
	},
	{
		Name:           "generate",
		HelpText:       "Generate a new address to receive coin.",
		ArgsUsage:      "(prefix)",
		RequiresClient: false,
		Handler: func(ctx *HandlerContext) error {
			var kc KeyConfig
			if len(ctx.Args) > 0 {
				if len(ctx.Args[0]) >= 6 {
					fmt.Println(yellowStr("longer prefixes take exponentially longer time to find"))
				} // nb. we don't check that the given prefix only contains hex chars
				for {
					priv, err := core.NewEcdsa()
					if err != nil {
						return err
					}
					tryKc := NewKeyConfig(priv)
					raw := tryKc.PublicKeyHash.Data()
					pkhHex := make([]byte, 64)
					hex.Encode(pkhHex, raw[:])
					if bytes.HasPrefix(pkhHex, []byte(ctx.Args[0])) {
						kc = tryKc
						break
					}
				}
			} else {
				priv, err := core.NewEcdsa()
				if err != nil {
					return err
				}
				kc = NewKeyConfig(priv)
			}
			ctx.Config.AddKeys(kc)
			fmt.Println(kc.PublicKeyHash)
			return ctx.Config.Save()
		},
	},
	{
		Name:           "balance",
		HelpText:       "Get the total balance of our public key hashes, or given public key hashes.",
		ArgsUsage:      "(publicKeyHash...)",
		RequiredArgs:   0,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			var pkhs []core.HashT
			if len(ctx.Args) > 0 {
				pkhs = make([]core.HashT, len(ctx.Args))
				for i, arg := range ctx.Args {
					pkh, err := core.NewHashTFromString(arg)
					if err != nil {
						return err
					}
					pkhs[i] = pkh
				}
			} else {
				pkhs = ctx.Config.GetPublicKeyHashes()
			}
			// Actually get balances
			balances, err := ctx.Client.GetManyBalances(pkhs)
			if err != nil {
				return err
			}
			sort.Slice(pkhs, func(i, j int) bool {
				// > instead of < becaues we want descending
				return balances[pkhs[i]] > balances[pkhs[j]]
			})
			total := uint64(0)
			covered := set.NewSet[core.HashT]() // Don't consider duplicate pkhs
			for _, pkh := range pkhs {
				if !covered.Includes(pkh) {
					total += balances[pkh]
					covered.Add(pkh)
					fmt.Printf("%s\t%d\n", pkh, balances[pkh])
				}
			}
			fmt.Printf("\ntotal\t%d\n", total)
			return nil
		},
	},
	{
		Name:           "utxos",
		HelpText:       "Get the combined utxos of our public key hashes, or given public key hashes.",
		ArgsUsage:      "(publicKeyHash...)",
		RequiredArgs:   0,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			var pkhs []core.HashT
			if len(ctx.Args) > 0 {
				pkhs = make([]core.HashT, len(ctx.Args))
				for i, arg := range ctx.Args {
					pkh, err := core.NewHashTFromString(arg)
					if err != nil {
						return err
					}
					pkhs[i] = pkh
				}
			} else {
				pkhs = ctx.Config.GetPublicKeyHashes()
			}
			// Actually get utxos
			utxos, err := ctx.Client.GetManyUtxos(pkhs)
			if err != nil {
				return err
			}
			for utxo := range utxos {
				fmt.Printf("%s[%d]\t%d\n", utxo.TxId, utxo.Ind, utxo.Value)
			}
			return nil
		},
	},
}

func main() {
	Execute(commands)
}