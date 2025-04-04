package accounting

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type GroupByPriced struct {
	Opts   traverser.Options
	Values map[string]uint64
}

func (c *GroupByPriced) Traverse(r *types.Statement) {
	if len(c.Values) == 0 {
		c.Values = make(map[string]uint64)
	}
	c.Values[c.GetKey(r)]++
}

func (c *GroupByPriced) GetKey(r *types.Statement) string {
	status := "unpriced"
	if !r.SpotPrice.IsZero() {
		status = "priced"
	}
	return status + "_" + r.Asset.Hex() + "_" + r.Symbol + "," + c.Opts.Names[base.HexToAddress(r.Asset.String())].Name
}

func (c *GroupByPriced) Result() string {
	return c.Name() + "\n" + c.reportValues("Assets", c.Values)
}

func (c *GroupByPriced) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *GroupByPriced) Sort(array []*types.Statement) {
	// Nothing to do
}

func (c *GroupByPriced) reportValues(msg string, m map[string]uint64) string {
	type stats struct {
		Address base.Address
		Symbol  string
		Count   uint64
		Status  string
	}
	nTransfers := 0
	nPriced := 0

	arr := make([]stats, 0, len(m))
	for k, v := range m {
		parts := strings.Split(k, "_")
		arr = append(arr, stats{Count: v, Status: parts[0], Address: base.HexToAddress(parts[1]), Symbol: parts[2]})
		nTransfers += int(v)
		if parts[0] == "priced" {
			nPriced += int(v)
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Count == arr[j].Count {
			if arr[i].Address == arr[j].Address {
				return arr[i].Symbol < arr[j].Symbol
			}
			return arr[i].Address.LessThan(arr[j].Address)
		}
		return arr[i].Count > arr[j].Count
	})

	ret := fmt.Sprintf("Number of %s: %d\n", msg, len(c.Values))
	ret += fmt.Sprintf("Number of Transfers: %d\n", nTransfers)

	ret += ExportHeader("Priced Assets", nPriced)
	ret += "Count,Asset,Symbol,Name\n"
	for _, val := range arr {
		if val.Status == "priced" {
			ret += fmt.Sprintf("%d,%s,%s\n", val.Count, val.Address, val.Symbol)
		}
	}

	ret += ExportHeader("Unpriced Assets", nTransfers-nPriced)
	ret += "Count,Address,Symbol,Name\n"
	for _, val := range arr {
		if val.Status == "unpriced" {
			ret += fmt.Sprintf("%d,%s,%s\n", val.Count, val.Address, val.Symbol)
		}
	}

	return ret
}
