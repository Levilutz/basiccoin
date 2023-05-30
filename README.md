# basiccoin
Basic blockchain-based cryptocurrency (WIP)

In the style of early bitcoin, with some simplifications.

## Building

```bash
go build -o basiccoin ./src
```

## Running

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
