#!/usr/bin/env bash

echo
echo "--------------------------------------------------------------------------------"
echo "Processing $1..."

# chifra monitors --decache $1
# exit

chifra export --fmt csv --articulate --first_block 16422793 --last_block 16530247 --cache $1 >raw/txs/$1.csv
wc raw/txs/$1.csv

chifra export --fmt csv --articulate --first_block 16422793 --last_block 16530247 --accounting --statements $1 >raw/recons/$1.csv
wc raw/recons/$1.csv

chifra export --fmt csv --articulate --first_block 16422793 --last_block 16530247 --logs $1 >raw/logs/$1.csv
wc raw/logs/$1.csv

cat raw/recons/$1.csv | cut -f6 -d, | cut -f2 -d'"' | cut -f1 -d'-' | sort | uniq -c | sort -n

cat raw/logs/$1.csv | grep -v "{code:-32000 message:invalid" | grep -v blockNumber >>summary/all_logs.csv
cat raw/recons/$1.csv | \
    tr '`' ' ' | \
    awk -v FS='","' -v OFS='`' '{print $7,$8,$12,$6,$1,$2,$3,$4,$15,$22,$36,$35,$16,$17,$18,$19,$10,$11,$5,$9,$13,$14,$21,$20}' | \
    sed 's/`/\",\"/g' | \
    sed 's/^/\"/' | \
    sed 's/$/\"/' | \
    sed 's/UTC\",\"\"/UTC\",\"/g' | \
    grep -v "{code:-32" |  \
    grep -v assetAddr >>summary/all_recons.csv
