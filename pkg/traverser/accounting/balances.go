package accounting

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type AssetStatement struct {
	Opts   traverser.Options
	Values map[string]*types.Statement
}

func (c *AssetStatement) Traverse(r *types.Statement) {
	if len(c.Values) == 0 {
		c.Values = make(map[string]*types.Statement)
	}
	c.Values[c.GetKey(r)] = r
}

func (c *AssetStatement) GetKey(r *types.Statement) string {
	return r.Asset.Hex() + "_" + r.Symbol
}

func (c *AssetStatement) Result() string {
	return c.Name() + "\n" + c.reportValues("Assets", c.Values)
}

func (c *AssetStatement) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *AssetStatement) Sort(array []*types.Statement) {
	// Nothing to do
}

func (c *AssetStatement) reportValues(msg string, m map[string]*types.Statement) string {
	type stats struct {
		Address base.Address
		Symbol  string
		Balance string
		Recon   *types.Statement
	}
	hasPriced := 0
	hasNotPriced := 0
	zeroPriced := 0
	zeroNotPriced := 0

	arr := make([]stats, 0, len(m))
	for k, val := range m {
		parts := strings.Split(k, "_")
		stat := stats{Recon: val, Address: base.HexToAddress(parts[0]), Symbol: parts[1]}
		x := big.Float{}
		x.SetString(val.EndBal.Text(10))
		stat.Balance = ToFmtStr(c.Opts.Denom, val.Decimals, val.SpotPrice, &x)
		arr = append(arr, stat)
		hasUnits := !val.EndBal.IsZero()
		priced := !val.SpotPrice.IsZero()
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
		eb1 := arr[i].Recon.EndBal
		eb2 := arr[j].Recon.EndBal
		if eb1.Equal(&eb2) {
			return arr[i].Address.LessThan(arr[j].Address)
		}
		return eb1.LessThan(&eb2)
	})

	ret := fmt.Sprintf("Number of %s: %d\n", msg, len(c.Values))

	ret += ExportHeader("Non-Zero Units Priced", hasPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		hasUnits := !val.Recon.EndBal.IsZero()
		priced := !val.Recon.SpotPrice.IsZero()
		if hasUnits && priced {
			format := "%s,%s,%s,%s,%s,%s\n"
			ret += fmt.Sprintf(
				format,
				val.Recon.Date(),
				val.Address,
				val.Symbol,
				val.Recon.PriceSource,
				val.Recon.SpotPrice.String(),
				val.Balance,
			)
		}
	}

	ret += ExportHeader("Non-Zero Units Unpriced", hasNotPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		hasUnits := !val.Recon.EndBal.IsZero()
		priced := !val.Recon.SpotPrice.IsZero()
		if hasUnits && !priced {
			format := "%s,%s,%s,%s,%s,%s\n"
			ret += fmt.Sprintf(
				format,
				val.Recon.Date(),
				val.Address,
				val.Symbol,
				val.Recon.PriceSource,
				val.Recon.SpotPrice.String(),
				val.Balance,
			)
		}
	}

	ret += ExportHeader("Zero Units Priced", zeroPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		hasUnits := !val.Recon.EndBal.IsZero()
		priced := val.Recon.SpotPrice.GreaterThan(base.ZeroFloat)
		if !hasUnits && priced {
			format := "%s,%s,%s,%s,%s,%s\n"
			ret += fmt.Sprintf(
				format,
				val.Recon.Date(),
				val.Address,
				val.Symbol,
				val.Recon.PriceSource,
				val.Recon.SpotPrice.String(),
				val.Balance,
			)
		}
	}

	ret += ExportHeader("Zero Units Unpriced", zeroNotPriced)
	ret += "Date,Asset,Symbol,Price Source,Spot Price,Units,Usd\n"
	for _, val := range arr {
		hasUnits := !val.Recon.EndBal.IsZero()
		priced := !val.Recon.SpotPrice.IsZero()
		if !hasUnits && !priced {
			format := "%s,%s,%s,%s,%s,%s\n"
			ret += fmt.Sprintf(
				format,
				val.Recon.Date(),
				val.Address,
				val.Symbol,
				val.Recon.PriceSource,
				val.Recon.SpotPrice.String(),
				val.Balance,
			)
		}
	}

	return ret
}
