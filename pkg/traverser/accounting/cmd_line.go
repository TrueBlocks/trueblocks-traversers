package accounting

import (
	"fmt"
	"os"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

func GetTraversers(opts traverser.Options) []traverser.Traverser[*types.Statement] {
	ret := make([]traverser.Traverser[*types.Statement], 0)
	for _, a := range os.Args {
		if a == "accounting.counter" || a == "counters" {
			ret = append(ret, &Counter{Opts: opts})
		}
		if strings.Contains(a, "by_asset") || a == "counters" {
			ret = append(ret, &CountByAsset{Opts: opts})
		}
		if strings.Contains(a, "by_function") || a == "counters" {
			ret = append(ret, &CountByFunction{Opts: opts})
		}
		if strings.Contains(a, "by_priced") || a == "groups" {
			ret = append(ret, &GroupByPriced{Opts: opts})
		}
		// if strings.Contains(a, "by_name") || a == "groups" {
		// 	ret = append(ret, &GroupByNamed{Opts: opts})
		// }
		if strings.Contains(a, "by_address") || a == "groups" {
			ret = append(ret, &GroupByAddress{Opts: opts, Source: "senders"})
			ret = append(ret, &GroupByAddress{Opts: opts, Source: "recipients"})
			ret = append(ret, &GroupByAddress{Opts: opts, Source: "pairings"})
		}
		if a == "senders" {
			ret = append(ret, &GroupByAddress{Opts: opts, Source: "senders"})
		}
		if a == "recipients" {
			ret = append(ret, &GroupByAddress{Opts: opts, Source: "recipients"})
		}
		if a == "pairings" {
			ret = append(ret, &GroupByAddress{Opts: opts, Source: "pairings"})
		}
		if a == "identity" {
			ret = append(ret, &Identity{Opts: opts})
		}
		if a == "statements" {
			ret = append(ret, &AssetStatement{Opts: opts})
		}
		if a == "profit_and_loss" {
			ret = append(ret, &ProfitAndLoss{Opts: opts})
		}
		if a == "excel" {
			return append(ret, &Excel{Opts: opts})
		}
	}

	return ret
}

func ExportHeader(msg string, count int) string {
	return fmt.Sprintf("\nFn: %s: %d\n", msg, count)
}
