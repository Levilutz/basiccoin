#!/usr/bin/env bash

# Startup script for first run of coin.levilutz.com node.

nohup ./bcnode \
    --addr=coin.levilutz.com:21720 \
    --miners=1 \
    --payout=eeeeeed6bdacc4d88f6e07ba9070a3bcc1d1648cb8393ecb47bbe02235e48a5a \
    --http-wallet \
    --listen=true \
    > ~/basiccoin.out \
    2> ~/basiccoin.err &
