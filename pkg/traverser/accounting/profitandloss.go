package accounting

import (
	"math/big"
	"os"
	"strings"
	"text/tabwriter"

	//    "accounting/pkg/utils"
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/utils"
)

// --------------------------------
type ProfitAndLoss struct {
	Opts     traverser.Options
	LastDate string
	Ledgers  map[string]*types.Statement
	LastKey  string
	w        *tabwriter.Writer
}

func (c *ProfitAndLoss) Traverse(r *types.Statement) {
	if len(c.Ledgers) == 0 {
		c.w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ',', 0)
		if c.Ledgers == nil { // order matters
			c.ReportHeader(c.Opts.Verbose, r)
		}
		c.Ledgers = make(map[string]*types.Statement)
		c.LastKey = ""
	}

	if len(c.Opts.AddrFilters) > 0 && !c.Opts.AddrFilters[r.Asset] {
		return
	}

	// fmt.Println(r)
	key := c.GetKey(r)
	l := c.Ledgers[key]
	if l != nil {
		// We have this ledger, so first report on the current reconciliation...
		if c.Opts.Verbose > 0 {
			c.Report("Tx", colors.BrightCyan, r.SpotPrice, r)
		}
		// ...then accumulate it into the ledger
		c.UpdateLedger(key, r)

	} else {
		if c.LastKey != "" {
			c.Report("Summary", colors.BrightYellow, r.SpotPrice, c.Ledgers[c.LastKey])
			if colors.White != "" {
				fmt.Fprintln(c.w)
			}
		}
		if c.Opts.Verbose > 0 {
			c.Report("Tx", colors.BrightCyan, r.SpotPrice, r)
		}
		// Remember the current ledger
		c.Ledgers[key] = r
	}

	c.LastDate = base.GetDateKey(c.Opts.Period, r.DateTime())
	c.LastKey = key
}

func (c *ProfitAndLoss) GetKey(r *types.Statement) string {
	if c.Opts.Period == "blockly" {
		return fmt.Sprintf("%s-%08d", c.GetAsset(r), r.BlockNumber)
	}

	return fmt.Sprintf("%s-%s", c.GetAsset(r), base.GetDateKey(c.Opts.Period, r.DateTime()))
}

func (c *ProfitAndLoss) GetAsset(r *types.Statement) string {
	return fmt.Sprintf("%s-%s-%s", r.Asset.String(), r.Symbol, r.AccountedFor.String())
}

func (c *ProfitAndLoss) Result() string {
	return ""
}

func (a *ProfitAndLoss) Name() string {
	return colors.Green + reflect.TypeOf(a).Elem().String() + colors.Off
}

func (c *ProfitAndLoss) Sort(array []*types.Statement) {
	// Nothing to do
}

func ToFmtStrFloat(denom string, decimals base.Value, spot base.Float, x string) string {
	bigTotIn := big.Float{}
	bigTotIn.SetString(x)
	return ToFmtStr(denom, decimals, spot, &bigTotIn)
}

func ToFmtStr(denom string, decimals base.Value, spot base.Float, x *big.Float) string {
	v := *x
	switch denom {
	case "units":
		one := big.Float{}
		one.SetFloat64(1)
		v = *utils.PriceUsd(v.String(), decimals, &one)
		return v.Text('f', int(decimals))
	case "usd":
		sp := (big.Float)(spot)
		v = *utils.PriceUsd(v.String(), decimals, &sp)
		return v.Text('f', 6)
	case "wei":
		fallthrough
	default:
		return v.Text('f', 0)
	}
}

func (c *ProfitAndLoss) UpdateLedger(key string, r *types.Statement) {
	abs := r.AmountNet().Abs()
	if r.AmountNet().GreaterThan(base.ZeroWei) {
		current := c.Ledgers[key].TotalIn()
		current.Add(current, abs)
		c.Ledgers[key].AmountIn = *current
	} else if r.AmountNet().LessThan(base.ZeroWei) {
		current := c.Ledgers[key].TotalOut()
		current.Add(current, abs)
		c.Ledgers[key].AmountOut = *current
	}
	c.Ledgers[key].EndBal = r.EndBal
}

func Display(color string, a base.Address, aF *base.Address, verbose int, nMap map[base.Address]types.Name) string {
	n := nMap[base.HexToAddress(a.String())].Name
	n = strings.Replace(strings.Replace(n, ",", "", -1), "#", "", -1)
	if len(n) > 0 {
		n = "," + n
	} else {
		n = "," + colors.BrightBlack + a.String() + colors.Off
	}

	ad := a.String()
	if aF != nil && a == *aF {
		return colors.Green + ad + colors.Off + n
	}

	if verbose > 0 {
		return color + ad + colors.Off + n
	}

	if len(ad) < 15 {
		return color + ad + strings.Repeat(" ", 17-len(ad)) + colors.Off + n
	}
	return color + ad[0:8] + "..." + ad[len(ad)-6:] + colors.Off + n
}

func (c *ProfitAndLoss) Report(msg, color string, spot base.Float, r *types.Statement) {
	if len(c.Opts.AddrFilters) > 0 && !c.Opts.AddrFilters[r.Asset] {
		return
	}

	f := func(x big.Float) uint64 {
		v, _ := x.Uint64()
		return v
	}

	denom := c.Opts.Denom
	if denom == "usd" && r.SpotPrice.IsZero() {
		denom = "not-priced"
	}
	date := colors.Red + base.GetDateKey(c.Opts.Period, r.DateTime())
	if msg != "Summary" {
		date = colors.Red + base.GetDateKey("secondly", r.DateTime())
	}
	symbol := colors.Green + r.Symbol
	asset := color + Display(color, r.Asset, nil, c.Opts.Verbose, c.Opts.Names)
	sender := Display(color, r.Sender, &r.AccountedFor, c.Opts.Verbose, c.Opts.Names)
	recipient := Display(color, r.Recipient, &r.AccountedFor, c.Opts.Verbose, c.Opts.Names)
	var x big.Float
	x.SetString(r.BegBal.Text(10))
	beg := color + ToFmtStr(c.Opts.Denom, r.Decimals, spot, &x)
	x.SetString(r.AmountNet().Text(10))
	net := ToFmtStr(c.Opts.Denom, r.Decimals, spot, &x)
	x.SetString(r.EndBal.Text(10))
	end := ToFmtStr(c.Opts.Denom, r.Decimals, spot, &x)
	if f(x) == 0 {
		end = colors.BrightBlack + end
	}
	bigTotIn := big.Float{}
	bigTotIn.SetString(r.TotalIn().Text(10))
	totIn := ToFmtStrFloat(c.Opts.Denom, r.Decimals, spot, r.TotalIn().Text(10))
	gasOut := ToFmtStrFloat(c.Opts.Denom, r.Decimals, spot, r.GasOut.Text(10))
	totOutLessGas := ToFmtStrFloat(c.Opts.Denom, r.Decimals, spot, r.TotalOutLessGas().Text(10))
	sig := strings.Split(strings.Replace(strings.Replace(r.Signature(), "{name:", "", -1), "}", "", -1), "|")[0]
	if len(sig) == 0 {
		sig = r.Encoding()
	}
	sig = colors.White + sig
	if msg == "Summary" {
		sig = "---------------"
	}
	hash := color + r.TransactionHash.String()

	checks := map[bool]string{false: colors.Red + "x", true: colors.Green + "ok"}
	check := checks[r.Reconciled()]

	var row string
	if msg == "Summary" {
		if c.Opts.Verbose > 1 {
			row = fmt.Sprintf(
				"%s\t\t\t\t\t%s\t%s\t%s\t%s\t\t\t\t\t\t\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\t\t%s%s",
				msg,
				date,
				r.AccountedFor,
				symbol,
				asset,
				r.Decimals,
				denom,
				beg,
				net,
				end,
				totIn,
				gasOut,
				totOutLessGas,
				check,
				colors.Off)
		} else {
			row = fmt.Sprintf(
				"%s\t\t\t%s\t%s\t%s\t\t\t\t\t\t\t%d\t%s\t%s\t%s\t%s\t\t\t%s%s",
				msg,
				date,
				symbol,
				asset,
				r.Decimals,
				denom,
				beg,
				net,
				end,
				check,
				colors.Off)
		}
	} else {
		if c.Opts.Verbose > 1 {
			row = fmt.Sprintf(
				"%s\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s%s",
				msg,
				r.BlockNumber,
				r.TransactionIndex,
				r.LogIndex,
				hash,
				date,
				r.AccountedFor,
				symbol,
				asset,
				sender,
				recipient,
				r.PriceSource,
				r.SpotPrice.String(),
				r.Decimals,
				denom,
				beg,
				net,
				end,
				totIn,
				gasOut,
				totOutLessGas,
				sig,
				r.ReconciliationType(),
				check,
				colors.Off)
		} else {
			row = fmt.Sprintf(
				"%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s%s",
				msg,
				r.BlockNumber,
				r.TransactionIndex,
				date,
				symbol,
				asset,
				sender,
				recipient,
				r.PriceSource,
				r.SpotPrice.String(),
				r.Decimals,
				denom,
				beg,
				net,
				end,
				sig,
				r.ReconciliationType(),
				check,
				colors.Off)
		}
	}
	fmt.Fprintf(c.w, "%s\n", row)
	c.w.Flush()
}

func (c *ProfitAndLoss) ReportHeader(verbose int, r *types.Statement) {
	var fields []string
	if verbose > 1 {
		fields = []string{
			"type",
			"blockNumber",
			"transactionIndex",
			"logIndex",
			"transactionHash",
			"date",
			"accountedFor",
			"assetSymbol",
			"assetAddress",
			"assetName",
			"sender",
			"senderName",
			"recipient",
			"recipientName",
			"priceSource",
			"spotPrice",
			"decimals",
			"denom",
			"begBal",
			"amountNet",
			"endBal",
			"totalIn",
			"gasOut",
			"totalOutLessGas",
			"function",
			"reconciliationType",
			"reconciled",
		}
	} else {
		fields = []string{
			"type",
			"blockNumber",
			"transactionIndex",
			"date",
			"assetSymbol",
			"assetAddress",
			"assetName",
			"sender",
			"senderName",
			"recipient",
			"recipientName",
			"priceSource",
			"spotPrice",
			"decimals",
			"denom",
			"begBal",
			"amountNet",
			"endBal",
			"function",
			"reconciliationType",
			"reconciled",
		}
	}
	for i, f := range fields {
		if i > 0 {
			fmt.Fprint(c.w, ",")
		}
		fmt.Fprint(c.w, f)
	}
	c.w.Write([]byte{'\n'})
	c.w.Flush()
}
