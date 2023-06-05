# basiccoin
Basic blockchain-based cryptocurrency (WIP)

In the style of early bitcoin, with some simplifications.

## Running a full node

Build
```bash
go build -o basiccoin ./src/fullnode
```

To start a new chain
```bash
./basiccoin --addr="<addr:port to host from>"
```

To connect to an existing chain
```bash
./basiccoin --seed="<addr:port of seed peer>"
```

For more info
```bash
./basiccoin --help
```

## Using the cli with an existing full node

Build
```bash
go build -o basiccoin-cli ./src/cli
```

Set up a new local wallet
```bash
./basiccoin-cli setup
```

Import an existing local wallet
```bash
./basiccoin-cli import [path]
```
