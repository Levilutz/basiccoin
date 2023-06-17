#!/usr/bin/env bash

set -ex

go build -o bcoin-node ./src/fullnode

go build -o bcoin ./src/cli
