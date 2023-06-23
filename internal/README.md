# Internal Node Components

The independent components of a basiccoin node.

## `bus`
The main shared pub-sub message bus. Broadcasts events, commands, and queries between components.

## `chain`
A local blockchain instance.

## `inv`
The shared inventory of known blocks, merkle nodes, and transactions.

## `miner`
A single-threaded miner.

## `peer`
A routine which manages a connection to a single peer.

## `peerfactory`
The peer factory listens for inbound connections, seeks new peers when appropriate, and tracks how many peers we have.

## `rest`
The wallet / node management HTTP server.
