# Internal Node Components

The independent components of a basiccoin node.

## `chain`
A local blockchain instance.

## `ctlapi`
An instance of the node management HTTP server.

## `miner`
A single-threaded miner.

## `peer`
A routine which manages a connection to a single peer.

## `peerfactory`
The peer factory listens for inbound connections, seeks new peers when appropriate, and tracks how many peers we have.

## `pubsub`
The main pub-sub event bus all components share.

## `walletapi`
An instance of the wallet HTTP server.
