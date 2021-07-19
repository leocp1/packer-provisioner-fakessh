#! /usr/bin/env bash

echo "Checking sorted.txt actually on remote"
[ ! -e sorted.txt ] && exit 1
[ "$(cat sorted.txt)" != "$(printf '1\n2\n3')" ] && exit 1

exit 0
