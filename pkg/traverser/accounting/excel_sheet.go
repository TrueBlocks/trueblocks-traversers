package accounting

import "github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"

type AssetSheet struct {
	Name      string
	RangeName string
	Address   string
	Symbol    string
	Decimals  int
	nRecords  int
	Records   []*types.Statement
}

func (sheet *AssetSheet) monthSwitches(txIndex int) bool {
	if txIndex == 0 {
		return false
	}
	pMonth := sheet.Records[txIndex-1].DateTime().Format("2006-01")
	cMonth := sheet.Records[txIndex].DateTime().Format("2006-01")
	return pMonth != cMonth
}

func (sheet *AssetSheet) yearSwitches(txIndex int) bool {
	if txIndex == 0 {
		return false
	}
	pYear := sheet.Records[txIndex-1].DateTime().Format("2006")
	cYear := sheet.Records[txIndex].DateTime().Format("2006")
	return pYear != cYear
}
