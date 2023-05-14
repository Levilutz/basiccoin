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
./basiccoin
```

To connect to an existing chain
```bash
./basiccoin --seed="<Peer's Address>"
```
