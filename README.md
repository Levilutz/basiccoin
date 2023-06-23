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
./bcnode
```

To start a mining node
```bash
./bcnode --miners <numCpuCores> --payout <publicKeyHash>
```
You can generate a miner payout address with the cli `generate` command. Your miner will probably find a few useless blocks while it's still syncing its chain with the seed peer.

For more info
```bash
./bcnode --help
```

### Using the cli to manage a wallet

To view available commands
```bash
./bcwallet help
```

To generate a new wallet address
```bash
./bcwallet generate
```

To view your balance
```bash
./bcwallet balance
```

To send money to an address
```bash
./bcwallet send <address>:<amount>
```
