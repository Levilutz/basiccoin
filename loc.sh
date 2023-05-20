#!/usr/bin/env bash

# We must minimize the loc

find . -iname "*.go" | grep -v "_test.go" | xargs wc -l
