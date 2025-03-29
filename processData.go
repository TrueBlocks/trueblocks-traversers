package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	sdk "github.com/TrueBlocks/trueblocks-sdk/v5"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/accounting"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/logs"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser/stats"
)

func processData() {
	opts := traverser.GetOptions()
	statTraversers := stats.GetTraversers(opts)
	reconTraversers := accounting.GetTraversers(opts)
	logTraversers := logs.GetTraversers(opts)
	sorted := map[string]bool{}

	statements, err := getStatements(&opts)
	if err != nil {
		log.Fatalf("Error in getStatements: %v", err)
	}
	for _, stmt := range statements {
		for _, a := range statTraversers {
			a.Traverse(float64(stmt.BlockNumber))
		}
		for _, a := range reconTraversers {
			if !sorted[a.Name()] {
				a.Sort(statements)
				sorted[a.Name()] = true
			}
			a.Traverse(stmt)
		}
	}

	logs, err := getLogs(&opts)
	if err != nil {
		log.Fatalf("Error in getLogs: %v", err)
	}
	for _, l := range logs {
		for _, a := range logTraversers {
			if !sorted[a.Name()] {
				a.Sort(logs)
				sorted[a.Name()] = true
			}
			a.Traverse(l)
		}
	}

	for _, a := range statTraversers {
		fmt.Println(a.Result())
	}

	for _, a := range reconTraversers {
		fmt.Println(a.Result())
	}

	for _, a := range logTraversers {
		fmt.Println(a.Result())
	}
}

func getStatements(opts *traverser.Options) ([]*types.Statement, error) {
	ret := make([]*types.Statement, 0, 2000)

	for _, account := range opts.Accounts {
		if !isOfInterest(account.Tags) {
			continue
		}
		log.Println(colors.Yellow+"Fetching statements for", account.Address.Hex(), account.Tags, account.Name, colors.Off)
		exportOptions := sdk.ExportOptions{
			Addrs:      []string{account.Address.Hex()},
			Articulate: true,
			Globals: sdk.Globals{
				Cache: true,
			},
		}
		if statements, _, err := exportOptions.ExportStatements(); err != nil {
			log.Fatalf("Error in export statements: %v", err)
			return nil, err
		} else {
			for _, s := range statements {
				ret = append(ret, &s)
			}
		}
	}

	log.Println(colors.Yellow+"Loaded", len(ret), "statements", colors.Off)
	return ret, nil
}

func getLogs(opts *traverser.Options) ([]*types.Log, error) {
	ret := make([]*types.Log, 0, 100)

	for _, account := range opts.Accounts {
		if !isOfInterest(account.Tags) {
			continue
		}
		log.Println(colors.Yellow+"Fetching logs for", account.Address.Hex(), account.Name, colors.Off)
		exportOptions := sdk.ExportOptions{
			Addrs:      []string{account.Address.Hex()},
			Articulate: true,
			Globals: sdk.Globals{
				Cache: true,
			},
		}
		logs, _, err := exportOptions.ExportLogs()
		if err != nil {
			log.Fatalf("Error in export logs: %v", err)
		}
		for _, l := range logs {
			ret = append(ret, &l)
		}
	}

	log.Println(colors.Yellow+"Loaded", len(ret), "logs", colors.Off)
	return ret, nil
}

func isOfInterest(tag string) bool {
	tags := []string{"00-Active", "11-Retired", "12-Empty", "14-Other", "17-Unused", "19-Dead"}
	return slices.Contains(tags, tag)
}
