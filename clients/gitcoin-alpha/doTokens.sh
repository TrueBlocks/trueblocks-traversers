#!/usr/bin/env bash

echo
echo "--------------------------------------------------------------------------------"
echo "Processing $1..."

# chifra monitors --decache $1

# make data
chifra export --fmt csv --articulate --logs $1 --emitter 0x6b175474e89094c44da98b954eedeac495271d0f --cache  --first_block 16422793 --last_block 16530247  >raw/logs/$1.csv

# group data
cat raw/logs/$1.csv | grep -v blockNumber >>summary/all_logs.csv

# show counts
echo -n "logs: "
wc -l raw/logs/$1.csv
