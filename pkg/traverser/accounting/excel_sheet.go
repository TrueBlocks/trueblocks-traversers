package accounting

import (
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
)

type AssetSheet struct {
	Name      string
	RangeName string
	Address   string
	Symbol    string
	Decimals  int
	nRecords  int
	Records   []*mytypes.RawReconciliation
}

func (sheet *AssetSheet) monthSwitches(txIndex int) bool {
	if txIndex == 0 {
		return false
	}
	pMonth := sheet.Records[txIndex-1].Date.Format("2006-01")
	cMonth := sheet.Records[txIndex].Date.Format("2006-01")
	return pMonth != cMonth
}

func (sheet *AssetSheet) yearSwitches(txIndex int) bool {
	if txIndex == 0 {
		return false
	}
	pYear := sheet.Records[txIndex-1].Date.Format("2006")
	cYear := sheet.Records[txIndex].Date.Format("2006")
	return pYear != cYear
}
