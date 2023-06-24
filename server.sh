#!/usr/bin/env bash
# Install and run a basiccoin node on a linux machine

set -ex

# Get curl and git
sudo apt-get update
sudo apt-get install -y curl git

# Get golang
curl -LO https://go.dev/dl/go1.20.5.linux-amd64.tar.gz
tar xvf go1.20.5.linux-amd64.tar.gz
sudo chown -R root:root ./go
sudo mv go /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
# shellcheck disable=SC1090
source ~/.bashrc

# Get basiccoin
git clone https://github.com/Levilutz/basiccoin.git
cd basiccoin
go build -o bcnode ./cmd/bcnode

# Run server
ufw allow 21720/tcp
ufw reload

nohup ./bcnode --new-network=true --listen=true --addr=coin1.levilutz.com:21720 --miners=1 --payout=eeeeeed6bdacc4d88f6e07ba9070a3bcc1d1648cb8393ecb47bbe02235e48a5a --http-wallet > ~/basiccoin.out 2> ~/basiccoin.err &
# nohup ./bcnode --listen=true --addr=coin2.levilutz.com:21720 --miners=1 --payout=eeeeeed6bdacc4d88f6e07ba9070a3bcc1d1648cb8393ecb47bbe02235e48a5a > ~/basiccoin.out 2> ~/basiccoin.err &
# nohup ./bcnode --listen=true --addr=coin3.levilutz.com:21720 --miners=1 --payout=eeeeeed6bdacc4d88f6e07ba9070a3bcc1d1648cb8393ecb47bbe02235e48a5a > ~/basiccoin.out 2> ~/basiccoin.err &

tail --follow ~/basiccoin.out
