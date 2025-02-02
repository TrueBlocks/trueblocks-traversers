#!/usr/bin/env bash

# chifra names -c $1
# exit

echo
echo "--------------------------------------------------------------------------------"
echo "Processing $1..."

yes | chifra export --decache $1
# exit

chifra export --fmt csv --articulate --cache --cache_traces $1 >raw/txs/$1.csv
wc raw/txs/$1.csv

chifra export --fmt csv --articulate --accounting --statements --cache --cache_traces $1 >raw/recons/$1.csv
# wc raw/recons/$1.csv

chifra export --fmt csv --articulate --logs $1 >raw/logs/$1.csv
# wc raw/logs/$1.csv

cat raw/recons/$1.csv | cut -f6 -d, | cut -f2 -d'"' | cut -f1 -d'-' | sort | uniq -c | sort -n

cat raw/logs/$1.csv | grep -v "{code:-32000 message:invalid" | grep -v blockNumber >>summary/all_logs.csv

cat raw/recons/$1.csv | \
    tr '`' ' ' | \
    awk -v FS=',' -v OFS='`' '{print $7,$8,$12,$6,$1,$2,$3,$4,$15,$20,$34,$33,$16,$17,"","",$10,$11,$5,$9,$13,$14,$19,$18}' | \
    sed 's/`/\",\"/g' | \
    sed 's/^/\"/' | \
    sed 's/$/\"/' | \
    sed 's/UTC\",\"\"/UTC\",\"/g' | \
    grep -v "{code:-32" |  \
    grep -v assetAddr >>summary/all_recons.csv
