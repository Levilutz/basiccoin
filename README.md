# basiccoin
A basic blockchain-based peer-to-peer cryptocurrency

## Requirements
* [golang](https://go.dev/) >= 1.20

## Building
```bash
bash build.sh
```

## Usage
### Running a full node

To start a node
```bash
./bcoin-node
```

To start a mining node
```bash
./bcoin-node --miners <num cpu cores> --minerPayoutAddr <address>
```
You can generate a miner payout address with the cli `generate` command.

For more info
```bash
./bcoin-node --help
```

### Using the cli to manage a wallet

To view available commands
```bash
./bcoin help
```

To generate a new wallet address
```bash
./bcoin generate
```

To view your balance
```bash
./bcoin balance
```

To send money to an address
```bash
./bcoin send <address> <amount>
```
