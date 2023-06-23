#!/usr/bin/env bash

set -ex

go build -o bcnode ./cmd/bcnode
go build -o bcwallet ./cmd/bcwallet
