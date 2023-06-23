package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

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
	{
		Name:           "send",
		HelpText:       "Send coin to given public key hashes, given as 'pkh:amount' pairs.",
		ArgsUsage:      "[pkh:amount...]",
		RequiredArgs:   1,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			// Parse input
			dests := make(map[core.HashT]uint64, len(ctx.Args))
			for _, arg := range ctx.Args {
				split := strings.Split(arg, ":")
				if len(split) != 2 {
					return fmt.Errorf("must provide 'pkh:amount' pairs")
				}
				pkh, err := core.NewHashTFromString(split[0])
				if err != nil {
					return err
				}
				val, err := strconv.ParseUint(split[1], 10, 64)
				if err != nil {
					return err
				}
				dests[pkh] = val
			}

			// TODO: Get min block
			// Get utxo balances
			utxos, err := ctx.Client.GetManyUtxos(ctx.Config.GetPublicKeyHashes())
			if err != nil {
				return err
			}

			// Make tx
			tx, err := core.MakeOutboundTx(
				ctx.Config.CoreParams(),
				ctx.Config.GetPrivateKeys(),
				utxos,
				dests,
				float64(1.0),
				0,
			)
			if err != nil {
				return err
			}

			// Ask user for confirmation on the fee rate
			fee := tx.InputsValue() - tx.OutputsValue()
			nonChangeOutputs := tx.OutputsValue() - tx.Outputs[0].Value
			feeRate := 100.0 * float64(fee) / float64(nonChangeOutputs)
			fmt.Printf("outputs: %d\n", nonChangeOutputs)
			fmt.Printf("fees: %d (%.2fp)\n", fee, feeRate)
			if inp := ReadInput("confirm? (y/n): "); inp != "y" && inp != "Y" {
				return fmt.Errorf("tx cancelled")
			}

			// Send tx
			resp, err := ctx.Client.PostTx(*tx)
			if err != nil {
				return err
			}

			fmt.Println(greenStr(resp.String()))
			return nil
		},
	},
}

func main() {
	Execute(commands)
}

// Read a line from stdin, given prompt.
func ReadInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return text[:len(text)-1]
}
