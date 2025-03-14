package accounting

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/utils"
)

// --------------------------------
type AssetStatement struct {
	Opts   traverser.Options
	Values map[string]*mytypes.RawReconciliation
}

func (c *AssetStatement) Traverse(r *mytypes.RawReconciliation) {
	if len(c.Values) == 0 {
		c.Values = make(map[string]*mytypes.RawReconciliation)
	}
	c.Values[c.GetKey(r)] = r
}

func (c *AssetStatement) GetKey(r *mytypes.RawReconciliation) string {
	return string(r.AssetAddress) + "_" + r.AssetSymbol
}

func (c *AssetStatement) Result() string {
	return c.Name() + "\n" + c.reportValues("Assets", c.Values)
}

func (c *AssetStatement) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *AssetStatement) Sort(array []*mytypes.RawReconciliation) {
	// Nothing to do
}

func (c *AssetStatement) reportValues(msg string, m map[string]*mytypes.RawReconciliation) string {
	type stats struct {
		Address string
		Symbol  string
		Balance string
		Recon   *mytypes.RawReconciliation
	}
	hasPriced := 0
	hasNotPriced := 0
	zeroPriced := 0
	zeroNotPriced := 0

	arr := make([]stats, 0, len(m))
	for k, v := range m {
		parts := strings.Split(k, "_")
		stat := stats{Recon: v, Address: parts[0], Symbol: parts[1]}
		var x big.Float
		x.SetString(v.EndBal)
		stat.Balance = ToFmtStr(c.Opts.Denom, v.Decimals, v.SpotPrice, &x)
		arr = append(arr, stat)
		hasUnits := x.Cmp(utils.Zero()) != 0
		priced := v.SpotPrice > 0
		if hasUnits && priced {
			hasPriced++
		} else if hasUnits && !priced {
			hasNotPriced++
		} else if !hasUnits && priced {
			zeroPriced++
		} else if !hasUnits && !priced {
			zeroNotPriced++
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		var xi big.Float
		xi.SetString(arr[i].Recon.EndBal)
		var xj big.Float
		xj.SetString(arr[j].Recon.EndBal)
		if xi.Cmp(&xj) == 0 {
			return arr[i].Address < arr[j].Address
		}
		return xi.Cmp(&xj) < 0
	})

	ret := fmt.Sprintf("Number of %s: %d\n", msg, len(c.Values))

	ret += ExportHeader("Non-Zero Units Priced", hasPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		var x big.Float
		x.SetString(val.Recon.EndBal)
		hasUnits := x.Cmp(utils.Zero()) != 0
		priced := val.Recon.SpotPrice > 0
		if hasUnits && priced {
			ret += fmt.Sprintf("%s,%s,%s,%s,%f,%s\n", val.Recon.Date.String(), val.Address, val.Symbol, val.Recon.PriceSource, val.Recon.SpotPrice, val.Balance)
		}
	}

	ret += ExportHeader("Non-Zero Units Unpriced", hasNotPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		var x big.Float
		x.SetString(val.Recon.EndBal)
		hasUnits := x.Cmp(utils.Zero()) != 0
		priced := val.Recon.SpotPrice > 0
		if hasUnits && !priced {
			ret += fmt.Sprintf("%s,%s,%s,%s,%f,%s\n", val.Recon.Date.String(), val.Address, val.Symbol, val.Recon.PriceSource, val.Recon.SpotPrice, val.Balance)
		}
	}

	ret += ExportHeader("Zero Units Priced", zeroPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		var x big.Float
		x.SetString(val.Recon.EndBal)
		hasUnits := x.Cmp(utils.Zero()) != 0
		priced := val.Recon.SpotPrice > 0
		if !hasUnits && priced {
			ret += fmt.Sprintf("%s,%s,%s,%s,%f,%s\n", val.Recon.Date.String(), val.Address, val.Symbol, val.Recon.PriceSource, val.Recon.SpotPrice, val.Balance)
		}
	}

	ret += ExportHeader("Zero Units Unpriced", zeroNotPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		var x big.Float
		x.SetString(val.Recon.EndBal)
		hasUnits := x.Cmp(utils.Zero()) != 0
		priced := val.Recon.SpotPrice > 0
		if !hasUnits && !priced {
			ret += fmt.Sprintf("%s,%s,%s,%s,%f,%s\n", val.Recon.Date.String(), val.Address, val.Symbol, val.Recon.PriceSource, val.Recon.SpotPrice, val.Balance)
		}
	}

	return ret
}
