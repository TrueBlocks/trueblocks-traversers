#!/usr/bin/env bash

echo
echo "--------------------------------------------------------------------------------"
echo "Processing $1..."

# chifra monitors --decache $1

# make data
chifra export --fmt csv --articulate --cache $1 --first_block 16422793 --last_block 16530247 >raw/txs/$1.csv
chifra export --fmt csv --articulate --accounting --statements $1  --first_block 16422793 --last_block 16530247 >raw/recons/$1.csv
chifra export --fmt csv --articulate --logs $1  --first_block 16422793 --last_block 16530247 >raw/logs/$1.csv

# group data
cat raw/recons/$1.csv | \
    sed 's/\",\"/\"@\"/g' | \
    sed 's/\\\"//g' | \
    tr ',' ' ' | \
    tr ' ' '_' | \
    tr '@' '\t' | \
    awk '{print $7,$8,$12,$6,$1,$2,$3,$4,$15,$16,$17,$18,$19,$10,$11,$5,$9,$13,$14,$21}' | \
    tr ' ' ',' | \
    tr '_' ' ' | \
    grep -v assetAddr >>summary/all_recons.csv
cat raw/logs/$1.csv | grep -v blockNumber >>summary/all_logs.csv

# show years
# cat raw/recons/$1.csv | cut -f6 -d, | cut -f2 -d'"' | cut -f1 -d'-' | sort | uniq -c | sort -n

# show counts
echo -n "txs: "
wc raw/txs/$1.csv

echo -n "recons: "
wc raw/recons/$1.csv

echo -n "logs: "
wc raw/logs/$1.csv
