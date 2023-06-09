package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/levilutz/basiccoin/internal/rest/client"
	"github.com/levilutz/basiccoin/pkg/core"
	"github.com/levilutz/basiccoin/pkg/set"
	"github.com/levilutz/basiccoin/pkg/util"
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
			if len(pkhs) == 0 {
				return fmt.Errorf("no publicKeyHashes in wallet - run 'bcwallet generate'")
			}
			// Actually get balances
			balances, err := ctx.Client.GetManyBalances(pkhs)
			if err != nil {
				return err
			}
			sort.Slice(pkhs, func(i, j int) bool {
				// descending
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
			utxos, err := ctx.Client.GetManyUtxos(pkhs, false)
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

			// Get current head height / min block
			minBlock, err := ctx.Client.GetHeadHeight()
			if err != nil {
				return err
			}

			// Get utxo balances
			utxos, err := ctx.Client.GetManyUtxos(ctx.Config.GetPublicKeyHashes(), true)
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
				minBlock,
			)
			if err != nil {
				return err
			}

			// Ask user for confirmation on the fee rate
			fee := tx.InputsValue() - tx.OutputsValue()
			nonChangeOutputs := tx.OutputsValue() - tx.Outputs[0].Value
			feeRate := 100.0 * float64(fee) / float64(nonChangeOutputs)
			fmt.Printf("outputs: %d\n", nonChangeOutputs)
			fmt.Printf("fees: %d : %.2f%%\n", fee, feeRate)
			fmt.Printf("total debit: %d\n", nonChangeOutputs+fee)
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
	{
		Name:           "consolidate",
		HelpText:       "Consolidate controlled balance into fewer utxos.",
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			// Get current head height / min block
			minBlock, err := ctx.Client.GetHeadHeight()
			if err != nil {
				return err
			}

			// Get utxo balances
			utxos, err := ctx.Client.GetManyUtxos(ctx.Config.GetPublicKeyHashes(), true)
			if err != nil {
				return err
			}

			// Make tx
			tx, err := core.MakeConsolidateTx(
				ctx.Config.CoreParams(),
				ctx.Config.GetPrivateKeys(),
				utxos,
				float64(1.0),
				minBlock,
			)
			if err != nil {
				return err
			}

			// Ask user for confirmation on the fee rate
			fee := tx.InputsValue() - tx.OutputsValue()
			feeRate := 100.0 * float64(fee) / float64(tx.InputsValue())
			fmt.Printf("inputs: %d\n", tx.InputsValue())
			fmt.Printf("outputs: %d\n", tx.OutputsValue())
			fmt.Printf("fees: %d : %.2f%%\n", fee, feeRate)
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
	{
		Name:           "tx-confirms",
		HelpText:       "Get the number of confirmations for given tx ids.",
		ArgsUsage:      "(txId...)",
		RequiredArgs:   1,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			txIds, err := core.UnmarshalHashTSlice(ctx.Args)
			if err != nil {
				return err
			}
			confirms, err := ctx.Client.GetTxConfirms(txIds)
			if err != nil {
				return err
			}
			knownTxIds := util.MapKeys(confirms)
			sort.Slice(knownTxIds, func(i, j int) bool {
				// descending
				return confirms[knownTxIds[i]] > confirms[knownTxIds[j]]
			})
			for _, txId := range knownTxIds {
				numStr := ""
				if confirms[txId] == 0 {
					numStr = yellowStr("0")
				} else {
					numStr = greenStr(fmt.Sprintf("%d", confirms[txId]))
				}
				fmt.Printf("%s\t%s\n", txId, numStr)
			}
			for _, txId := range txIds {
				if _, ok := confirms[txId]; !ok {
					uhOh := redStr("not known")
					fmt.Printf("%s\t%s\n", txId, uhOh)
				}
			}
			return nil
		},
	},
	{
		Name:           "tx-block",
		HelpText:       "Get the block each of the given tx ids was included in.",
		ArgsUsage:      "(txId...)",
		RequiredArgs:   1,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			txIds, err := core.UnmarshalHashTSlice(ctx.Args)
			if err != nil {
				return err
			}
			includedBlocks, err := ctx.Client.GetTxIncludedBlock(txIds)
			if err != nil {
				return err
			}
			for txId, blockId := range includedBlocks {
				fmt.Printf("%s\n\t%s\n", txId, greenStr(fmt.Sprint(blockId)))
			}
			for _, txId := range txIds {
				if _, ok := includedBlocks[txId]; !ok {
					uhOh := redStr("not included")
					fmt.Printf("%s\t%s\n", txId, uhOh)
				}
			}
			return nil
		},
	},
	{
		Name:           "get-tx",
		HelpText:       "Get the data for the given txs.",
		ArgsUsage:      "(txId...)",
		RequiredArgs:   1,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			txIds, err := core.UnmarshalHashTSlice(ctx.Args)
			if err != nil {
				return err
			}
			txs, err := ctx.Client.GetTx(txIds)
			if err != nil {
				return err
			}
			for txId, tx := range txs {
				fmt.Printf("%s:\n", greenStr(fmt.Sprint(txId)))
				out, err := json.MarshalIndent(tx, "", "    ")
				if err != nil {
					panic(err)
				}
				fmt.Println(string(out))
			}
			for _, txId := range txIds {
				if _, ok := txs[txId]; !ok {
					uhOh := redStr("not known")
					fmt.Printf("%s\t%s\n", txId, uhOh)
				}
			}
			return nil
		},
	},
	{
		Name:           "get-merkle",
		HelpText:       "Get the data for the given merkles.",
		ArgsUsage:      "(merkleId...)",
		RequiredArgs:   1,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			merkleIds, err := core.UnmarshalHashTSlice(ctx.Args)
			if err != nil {
				return err
			}
			merkles, err := ctx.Client.GetMerkle(merkleIds)
			if err != nil {
				return err
			}
			for merkleId, merkle := range merkles {
				fmt.Printf("%s:\n", greenStr(fmt.Sprint(merkleId)))
				out, err := json.MarshalIndent(merkle, "", "    ")
				if err != nil {
					panic(err)
				}
				fmt.Println(string(out))
			}
			for _, merkleId := range merkleIds {
				if _, ok := merkles[merkleId]; !ok {
					uhOh := redStr("not known")
					fmt.Printf("%s\t%s\n", merkleId, uhOh)
				}
			}
			return nil
		},
	},
	{
		Name:           "get-block",
		HelpText:       "Get the data for the given blocks.",
		ArgsUsage:      "(blockId...)",
		RequiredArgs:   1,
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			blockIds, err := core.UnmarshalHashTSlice(ctx.Args)
			if err != nil {
				return err
			}
			blocks, err := ctx.Client.GetBlock(blockIds)
			if err != nil {
				return err
			}
			for blockId, block := range blocks {
				fmt.Printf("%s:\n", greenStr(fmt.Sprint(blockId)))
				out, err := json.MarshalIndent(block, "", "    ")
				if err != nil {
					panic(err)
				}
				fmt.Println(string(out))
			}
			for _, blockId := range blockIds {
				if _, ok := blocks[blockId]; !ok {
					uhOh := redStr("not known")
					fmt.Printf("%s\t%s\n", blockId, uhOh)
				}
			}
			return nil
		},
	},
	{
		Name:           "rich-list",
		HelpText:       "Get the current wealthiest publicKeyHashes.",
		ArgsUsage:      "(length)",
		RequiresClient: true,
		Handler: func(ctx *HandlerContext) error {
			maxLen := uint64(10)
			var err error
			if len(ctx.Args) > 0 {
				maxLen, err = strconv.ParseUint(ctx.Args[0], 10, 64)
				if err != nil {
					return err
				}
			}
			richList, err := ctx.Client.GetRichList(maxLen)
			if err != nil {
				return err
			} else if len(richList) == 0 {
				return fmt.Errorf("no publicKeyHashes have balance")
			}
			pkhs := util.MapKeys(richList)
			sort.Slice(pkhs, func(i, j int) bool {
				return richList[pkhs[i]] > richList[pkhs[j]]
			})
			for _, pkh := range pkhs {
				fmt.Printf("%s\t%d\n", pkh, richList[pkh])
			}
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
