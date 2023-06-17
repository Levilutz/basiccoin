package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/levilutz/basiccoin/src/kern"
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
			_, err := NewClient(&Config{NodeAddr: addr})
			if err != nil {
				return fmt.Errorf("failed to connect to client: " + err.Error())
			}
			ctx.Config.NodeAddr = addr
			return ctx.Config.Save()
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
			priv, err := kern.NewEcdsa()
			if err != nil {
				return err
			}
			kc := NewKeyConfig(priv)
			ctx.Config.AddKeys(kc)
			fmt.Printf("%s\n", kc.PublicKeyHash)
			return ctx.Config.Save()
		},
	},
	{
		Name:           "balance",
		HelpText:       "Get the total balance of all currently controlled addresses, or given addresses.",
		ArgsUsage:      "(address...)",
		RequiredArgs:   0,
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			var pkhs []kern.HashT
			if len(ctx.Args) > 0 {
				// Get balancekern.HashTn addresses
				pkhs = make([]kern.HashT, len(ctx.Args))
				for i, arg := range ctx.Args {
					pkh, err := kern.NewHashTFromString(arg)
					if err != nil {
						return err
					}
					pkhs[i] = pkh
				}
			} else {
				// Get balance of controlled addresses
				pkhs = ctx.Config.GetPublicKeyHashes()
			}
			balanceData, err := ctx.Client.GetAllBalances(pkhs)
			if err != nil {
				return err
			}
			for _, pkh := range balanceData.SortedAddrs {
				fmt.Printf("%s\t%d\n", pkh, balanceData.Balances[pkh])
			}
			fmt.Printf("total\t%d\n", balanceData.Total)
			return nil
		},
	},
	{
		Name:           "utxos",
		HelpText:       "Get all utxos available to currently controlled addresses.",
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			utxos, err := ctx.Client.GetAllUtxos(ctx.Config.GetPublicKeyHashes())
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
		HelpText:       "Send coin to a given address.",
		ArgsUsage:      "[address] [amount]",
		RequiredArgs:   2,
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			destPkh, err := kern.NewHashTFromString(ctx.Args[0])
			if err != nil {
				return err
			}
			amt, err := strconv.ParseUint(ctx.Args[1], 10, 64)
			if err != nil {
				return err
			}
			outputValues := map[kern.HashT]uint64{
				destPkh: amt,
			}
			tx, err := ctx.Client.MakeOutboundTx(outputValues)
			if err != nil {
				return err
			}
			txId, err := ctx.Client.SendTx(tx)
			if err != nil {
				return err
			}
			fmt.Println(greenStr(fmt.Sprint(txId)))
			return nil
		},
	},
	{
		Name:           "consolidate",
		HelpText:       "Consolidate utxos to the wealthiest controlled address",
		RequiresClient: true,
		Handler: func(ctx HandlerContext) error {
			tx, err := ctx.Client.MakeConsolidateTx()
			if err != nil {
				return err
			}
			txId, err := ctx.Client.SendTx(tx)
			if err != nil {
				return err
			}
			fmt.Println(greenStr(fmt.Sprint(txId)))
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
