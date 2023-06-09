# basiccoin
Basic blockchain-based cryptocurrency (WIP)

In the style of early bitcoin, with some simplifications.

## Running a full node

Build
```bash
go build -o basiccoin ./src/fullnode
```

To start a node
```bash
./basiccoin
```

To start a mining node
```bash
./basiccoin --miners 8 --minerPayoutAddr <address>
```
You can generate a miner payout address with the cli `generate` command.

For more info
```bash
./basiccoin --help
```

## Using the cli with an existing full node

Build
```bash
go build -o basiccoin-cli ./src/cli
```

To generate an address
```bash
./basiccoin-cli generate
```

To view your balance
```bash
./basiccoin-cli balance
```

For more help
```bash
./basiccoin-cli help
```
