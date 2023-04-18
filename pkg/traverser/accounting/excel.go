package accounting

import (
	"fmt"
	"log"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/excel"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuri/excelize/v2"
)

// --------------------------------
type Excel struct {
	Opts      traverser.Options
	ExcelFile *excelize.File
	Line      int
	Assets    map[string][]*mytypes.RawReconciliation
	Ignored   map[string]bool
}

func (c *Excel) Traverse(r *mytypes.RawReconciliation) {
	if c.ExcelFile == nil {
		c.ExcelFile = excel.NewWorkbook("Summary", []string{"This is the summary text"})
		c.Assets = make(map[string][]*mytypes.RawReconciliation)
		c.Ignored = make(map[string]bool)
	}

	interesting := c.Opts.AddrFilters[r.AssetAddress]
	if len(c.Opts.AddrFilters) > 0 && !interesting {
		if !c.Ignored[r.AssetAddress.String()] {
			fmt.Println("Skipping asset", r.AssetAddress.String())
		}
		c.Ignored[r.AssetAddress.String()] = true
		return
	}

	// f := mytypes.NewDateTime2(2022, 1, 1, 0, 0, 1)
	// l := mytypes.NewDateTime2(2023, 1, 1, 0, 0, 1)
	// if r.Date.Before(&f) || r.Date.After(&l) {
	// 	return
	// }

	c.Line += 1
	c.Assets[c.GetKey(r)] = append(c.Assets[c.GetKey(r)], r)
}

func (c *Excel) GetKey(r *mytypes.RawReconciliation) string {
	return fmt.Sprintf("%s-%s-%d", r.AssetAddress.String(), r.AssetSymbol, r.Decimals)
}

type AssetSheet struct {
	Name      string
	RangeName string
	Address   string
	Symbol    string
	Decimals  int
	nRecords  int
	Records   []*mytypes.RawReconciliation
}

func (sheet *AssetSheet) MonthSwitch(txIndex int) bool {
	if txIndex == 0 {
		return false
	}
	pMonth := sheet.Records[txIndex-1].Date.Format("2006-01")
	cMonth := sheet.Records[txIndex].Date.Format("2006-01")
	return pMonth != cMonth
}

func (sheet *AssetSheet) YearSwitch(txIndex int) bool {
	if txIndex == 0 {
		return false
	}
	pYear := sheet.Records[txIndex-1].Date.Format("2006")
	cYear := sheet.Records[txIndex].Date.Format("2006")
	return pYear != cYear
}

type Field struct {
	Order   int
	Column  string
	Wid     float64
	Format  string
	Formula string
	Style   int
}

const headerRow = 6

type Styles struct {
	regular      int
	integer      int
	accounting2  int
	accounting5  int
	price        int
	dateYear     int
	dateMonth    int
	date         int
	boolean      int
	address1     int
	address2     int
	address3     int
	bigInteger   int
	zero         int
	link         int
	tableHeader  int
	mainHeader   int
	monthRow     int
	monthRowDate int
	monthRow2    int
	monthRow5    int
	yearRow      int
	yearRowDate  int
	yearRow2     int
	yearRow5     int
}

type Rng struct {
	A       int
	B       int
	pDate   mytypes.DateTime
	Rows    []int
	CurRows []int
}

func (c *Excel) Result() string {
	styles, err := c.GetStyles()
	if err != nil {
		panic(err)
	}

	var fieldMap = map[string]*Field{
		"Type":            {1, "A", 5, "string", "", styles.regular},
		"Bn":              {2, "B", 12, "int", "", styles.integer},
		"TxId":            {3, "C", 7, "int", "", styles.integer},
		"LogId":           {4, "D", 7, "int", "", styles.integer},
		"Year":            {5, "E", 0, "date", "", styles.dateYear},
		"Month":           {6, "F", 0, "date", "", styles.dateMonth},
		"Date":            {7, "G", 25, "date", "", styles.date},
		"PrevUsd":         {8, "H", 15, "formula", "=N{R-1}", styles.accounting2},
		"ChangeUsd":       {9, "I", 15, "formula", "=J{R}-H{R}", styles.accounting2},
		"BegUsd":          {10, "J", 15, "formula", "=O{R}*Q{R}", styles.accounting2},
		"InUsd":           {11, "K", 15, "formula", "=O{R}*R{R}", styles.accounting2},
		"OutUsd":          {12, "L", 15, "formula", "=O{R}*S{R}", styles.accounting2},
		"GasUsd":          {13, "M", 15, "formula", "=O{R}*T{R}", styles.accounting2},
		"EndUsd":          {14, "N", 15, "formula", "=O{R}*U{R}", styles.accounting2},
		"Spot":            {15, "O", 15, "float2", "", styles.price},
		"Source":          {16, "P", 15, "string", "", styles.regular},
		"BegUnits":        {17, "Q", 18, "float5", "", styles.accounting5},
		"InUnits":         {18, "R", 18, "float5", "", styles.accounting5},
		"OutUnits":        {19, "S", 18, "float5", "", styles.accounting5},
		"GasUnits":        {20, "T", 18, "float5", "", styles.accounting5},
		"EndUnits":        {21, "U", 18, "float5", "", styles.accounting5},
		"BegBal":          {22, "V", 0, "big", "", styles.bigInteger},
		"Inflow":          {23, "W", 0, "big", "", styles.bigInteger},
		"Outflow":         {24, "X", 0, "big", "", styles.bigInteger},
		"GasOut":          {25, "Y", 0, "big", "", styles.bigInteger},
		"EndBal":          {26, "Z", 0, "big", "", styles.bigInteger},
		"Check":           {27, "AA", 18, "formula", "=ROUND(Q{R}+R{R}-S{R}-T{R}-U{R},5)", styles.accounting5},
		"Message":         {30, "AB", 20, "string", "", styles.regular},
		"ReconType":       {31, "AC", 0, "string", "", styles.regular},
		"Sender":          {32, "AD", 18, "address", "", styles.address1},
		"Recipient":       {33, "AE", 18, "address", "", styles.address1},
		"AccountedFor":    {34, "AF", 52, "address", "", styles.address1},
		"TransactionHash": {35, "AG", 80, "hash", "", styles.address1},
	}
	lastCol := "AG"

	var monthSummary = map[string]*Field{
		"Date":      {1, "G", 25, "date", "", styles.date},
		"PrevUsd":   {2, "H", 15, "formula", "=H{A}", styles.monthRow2},
		"ChangeUsd": {3, "I", 15, "formula", "=SUM(I{A}:I{B})", styles.monthRow2},
		"BegUsd":    {4, "J", 15, "formula", "=J{A}", styles.monthRow2},
		"InUsd":     {5, "K", 15, "formula", "=SUM(K{A}:K{B})", styles.monthRow2},
		"OutUsd":    {6, "L", 15, "formula", "=SUM(L{A}:L{B})", styles.monthRow2},
		"GasUsd":    {7, "M", 15, "formula", "=SUM(M{A}:M{B})", styles.monthRow2},
		"EndUsd":    {8, "N", 15, "formula", "=N{R-1}", styles.monthRow2},
		"CheckUsd":  {9, "O", 15, "formula", "=H{R}+I{R}+K{R}-L{R}-M{R}-N{R}", styles.price},
		"BegUnits":  {10, "Q", 18, "formula", "=Q{A}", styles.monthRow5},
		"InUnits":   {11, "R", 18, "formula", "=SUM(R{A}:R{B})", styles.monthRow5},
		"OutUnits":  {12, "S", 18, "formula", "=SUM(S{A}:S{B})", styles.monthRow5},
		"GasUnits":  {13, "T", 18, "formula", "=SUM(T{A}:T{B})", styles.monthRow5},
		"EndUnits":  {14, "U", 18, "formula", "=U{R-1}", styles.monthRow5},
		"Check":     {15, "AA", 15, "formula", "=Q{R}+R{R}-S{R}-T{R}-U{R}", styles.price},
	}

	var yearSummary = map[string]*Field{
		"Date":      {1, "G", 25, "date", "", styles.date},
		"PrevUsd":   {2, "H", 15, "formula", "=H{A}", styles.monthRow2},
		"ChangeUsd": {3, "I", 15, "formula", "=SUM(I{A}:I{B})", styles.monthRow2},
		"BegUsd":    {4, "J", 15, "formula", "=J{A}", styles.monthRow2},
		"InUsd":     {5, "K", 15, "formula", "=SUM(K{A}:K{B})", styles.monthRow2},
		"OutUsd":    {6, "L", 15, "formula", "=SUM(L{A}:L{B})", styles.monthRow2},
		"GasUsd":    {7, "M", 15, "formula", "=SUM(M{A}:M{B})", styles.monthRow2},
		"EndUsd":    {8, "N", 15, "formula", "=N{R-1}", styles.monthRow2},
		"CheckUsd":  {9, "O", 15, "formula", "=H{R}+I{R}+K{R}-L{R}-M{R}-N{R}", styles.price},
		"BegUnits":  {10, "Q", 18, "formula", "=Q{A}", styles.monthRow5},
		"InUnits":   {11, "R", 18, "formula", "=SUM(R{A}:R{B})", styles.monthRow5},
		"OutUnits":  {12, "S", 18, "formula", "=SUM(S{A}:S{B})", styles.monthRow5},
		"GasUnits":  {13, "T", 18, "formula", "=SUM(T{A}:T{B})", styles.monthRow5},
		"EndUnits":  {14, "U", 18, "formula", "=U{R-1}", styles.monthRow5},
		"Check":     {15, "AA", 15, "formula", "=Q{R}+R{R}-S{R}-T{R}-U{R}", styles.price},
	}

	type sorter struct {
		Order int
		Name  string
	}

	sortedFields := []sorter{}
	for k, v := range fieldMap {
		sortedFields = append(sortedFields, sorter{Order: v.Order, Name: k})
	}
	sort.Slice(sortedFields, func(i, j int) bool {
		return sortedFields[i].Order < sortedFields[j].Order
	})
	fields := []string{}
	for _, v := range sortedFields {
		fields = append(fields, v.Name)
	}

	sheets := c.ConvertToSheets()
	for _, sheet := range sheets {
		c.ExcelFile.NewSheet(sheet.Name)
		c.SetHeader(&sheet, &styles, lastCol)
		for _, field := range fields {
			c.ExcelFile.SetColWidth(sheet.Name, fieldMap[field].Column, fieldMap[field].Column, fieldMap[field].Wid)
		}

		bottomRow := headerRow
		verbose := true

		c.ExcelFile.SetSheetRow(sheet.Name, fmt.Sprintf("A%d", headerRow), &fields)

		rowRange := Rng{}
		monthRange := Rng{
			A: headerRow,
			B: headerRow,
		}
		yearRange := Rng{
			A: headerRow,
			B: headerRow,
		}

		lastRowType := ""
		curRow := headerRow
		for txIndex, r := range sheet.Records {
			sig := r.Signature
			if sig == "" {
				sig = r.Encoding
			} else {
				parts := strings.Split(sig, "|")
				sig = strings.Replace(strings.Replace(parts[0], "{name:", "", -1), "}", "", -1)
			}

			toUnits := func(v string, decimals int) float64 {
				var x big.Float
				x.SetString(v)
				val, _ := ToUnits(&x, r.Decimals).Float64()
				return val
			}

			begUnits := toUnits(r.BegBal, int(r.Decimals))
			inUnits := toUnits(r.TotalIn, int(r.Decimals))
			outUnitsLessGas := toUnits(r.TotalOutLessGas, int(r.Decimals))
			gasUnitsOut := toUnits(r.GasOut, int(r.Decimals))
			endUnits := toUnits(r.EndBal, int(r.Decimals))

			if sheet.MonthSwitch(txIndex) {
				if lastRowType == "Tx" || lastRowType == "" {
					curRow++
					monthRange.B = curRow
					monthRange.Rows = append(monthRange.Rows, curRow)
					monthRange.CurRows = append(monthRange.CurRows, curRow)
					msg := "Mo"
					if verbose {
						msg = fmt.Sprintf("%s-[%03d] [%03d-%03d]", "Mo", curRow, monthRange.A+1, monthRange.B-1)
					}
					c.SetCell(sheet.Name, curRow, monthRange, fieldMap["Type"], msg)
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["Date"], monthRange.pDate)
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["PrevUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["ChangeUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["BegUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["InUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["OutUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["GasUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["EndUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["CheckUsd"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["BegUnits"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["InUnits"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["OutUnits"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["GasUnits"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["EndUnits"], "")
					c.SetCell(sheet.Name, curRow, monthRange, monthSummary["Check"], "")

					monthRange.A = curRow
					lastRowType = "Mo"
					if sheet.YearSwitch(txIndex) {
						curRow++
						yearRange.Rows = append(yearRange.Rows, curRow)
						yearRange.CurRows = append(yearRange.CurRows, curRow)
						toStr := func(rows []int) string {
							str := ""
							for _, row := range rows {
								str += fmt.Sprintf("%d,", row)
							}
							return str
						}
						msg := "Yr"
						if verbose {
							msg = fmt.Sprintf("%s-[%03d] [%s]", "Yr", curRow, toStr(monthRange.CurRows))
						}
						c.SetCell(sheet.Name, curRow, yearRange, fieldMap["Type"], msg)
						c.SetCell(sheet.Name, curRow, yearRange, yearSummary["Date"], yearRange.pDate)
						monthRange.CurRows = []int{}
						monthRange.A = curRow
						lastRowType = "Yr"
					}
				}
			}

			if true {
				curRow++
				lastRowType = "Tx"
				chUsd := fieldMap["ChangeUsd"]
				if curRow == headerRow+1 {
					chUsd.Formula = "=0"
				} else {
					chUsd.Formula = "=J{R}-H{R}"
				}
				pUsd := fieldMap["PrevUsd"]
				if curRow == headerRow+1 {
					pUsd.Formula = "=0"
				} else {
					pUsd.Formula = "=N{R-1}"
				}

				// c.SetCell(sheet.Name, curRow, fieldMap["Type"], fmt.Sprintf("%s-[%03d]", "Tx", curRow))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Type"], fmt.Sprintf("%s-%d", "Tx", txIndex))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Bn"], int(r.BlockNumber))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["TxId"], int(r.TransactionIndex))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["LogId"], int(r.LogIndex))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Year"], r.Date.EndOfYear())
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Month"], r.Date.EndOfMonth())
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Date"], r.Date)
				c.SetCell(sheet.Name, curRow, rowRange, pUsd, "prevEnd")
				c.SetCell(sheet.Name, curRow, rowRange, chUsd, "(r.SpotPrice*begUnits)-prevEnd")
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["BegUsd"], r.SpotPrice*begUnits)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["InUsd"], r.SpotPrice*inUnits)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["OutUsd"], r.SpotPrice*outUnitsLessGas)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["GasUsd"], r.SpotPrice*gasUnitsOut)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["EndUsd"], r.SpotPrice*endUnits)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Spot"], r.SpotPrice)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Source"], r.PriceSource)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["BegUnits"], begUnits)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["InUnits"], inUnits)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["OutUnits"], outUnitsLessGas)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["GasUnits"], gasUnitsOut)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["EndUnits"], endUnits)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["BegBal"], r.BegBal)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Inflow"], r.TotalIn)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Outflow"], r.TotalOutLessGas)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["GasOut"], r.GasOut)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["EndBal"], r.EndBal)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Check"], "check")
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Message"], sig)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["ReconType"], r.ReconciliationType)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Sender"], r.Sender)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Recipient"], r.Recipient)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["AccountedFor"], r.AccountedFor)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["TransactionHash"], r.TransactionHash)

				// both or neither can be true...
				senderCell := fmt.Sprintf("%s%d", fieldMap["Sender"].Column, curRow)
				if r.Sender.IsZero() {
					c.SetStyle(sheet.Name, senderCell, senderCell, styles.zero)
				} else {
					style := styles.address1
					if c.Opts.Names[common.HexToAddress(r.Sender.String())].IsCustom {
						style = styles.address3
					}
					if r.Sender == r.AccountedFor {
						style = styles.address2
					}
					c.SetStyle(sheet.Name, senderCell, senderCell, style)
				}

				recipCell := fmt.Sprintf("%s%d", fieldMap["Recipient"].Column, curRow)
				if r.Recipient.IsZero() {
					c.SetStyle(sheet.Name, recipCell, recipCell, styles.zero)
				} else {
					style := styles.address1
					if c.Opts.Names[common.HexToAddress(r.Recipient.String())].IsCustom {
						style = styles.address3
					}
					if r.Recipient == r.AccountedFor {
						style = styles.address2
					}
					c.SetStyle(sheet.Name, recipCell, recipCell, style)
				}
			}

			bottomRow = curRow
			// prevEnd = r.SpotPrice * endUnits
			monthRange.pDate = r.Date
			yearRange.pDate = r.Date
		}

		if lastRowType == "Tx" || lastRowType == "" {
			curRow++
			monthRange.B = curRow
			monthRange.Rows = append(monthRange.Rows, curRow)
			monthRange.CurRows = append(monthRange.CurRows, curRow)
			msg := "Mo"
			if verbose {
				msg = fmt.Sprintf("%s-[%03d] [%03d-%03d]", "Mo", curRow, monthRange.A+1, monthRange.B-1)
			}
			c.SetCell(sheet.Name, curRow, monthRange, fieldMap["Type"], msg)
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["Date"], monthRange.pDate)
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["PrevUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["ChangeUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["BegUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["InUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["OutUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["GasUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["EndUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["CheckUsd"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["BegUnits"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["InUnits"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["OutUnits"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["GasUnits"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["EndUnits"], "")
			c.SetCell(sheet.Name, curRow, monthRange, monthSummary["Check"], "")
			monthRange.A = curRow
			lastRowType = "Mo"
		}

		if lastRowType == "Mo" || lastRowType == "" {
			curRow++
			yearRange.Rows = append(yearRange.Rows, curRow)
			yearRange.CurRows = append(yearRange.CurRows, curRow)
			toStr := func(rows []int) string {
				str := ""
				for _, row := range rows {
					str += fmt.Sprintf("%d,", row)
				}
				return str
			}
			msg := "Yr"
			if verbose {
				msg = fmt.Sprintf("%s-[%03d] [%s]", "Yr", curRow, toStr(monthRange.CurRows))
			}
			c.SetCell(sheet.Name, curRow, yearRange, fieldMap["Type"], msg)
			c.SetCell(sheet.Name, curRow, yearRange, yearSummary["Date"], yearRange.pDate)
			monthRange.CurRows = []int{}
			monthRange.A = curRow
			lastRowType = "Yr"
			curRow++
		}

		for _, field := range fields {
			if field == "Sender" || field == "Recipient" {
				continue
			}
			cell1 := fmt.Sprintf("%s%d", fieldMap[field].Column, headerRow+1)
			cell2 := fmt.Sprintf("%s%d", fieldMap[field].Column, bottomRow)
			style := fieldMap[field].Style
			c.SetStyle(sheet.Name, cell1, cell2, style)
		}

		// showStripes := true
		// cellRange := fmt.Sprintf("%s%d:%s%d", fieldMap[fields[0]].Column, headerRow, fieldMap[fields[len(fields)-1]].Column, bottomRow)
		// err = c.ExcelFile.AddTable(sheet.Name, cellRange, &excelize.TableOptions{
		// 	Name:              sheet.RangeName,
		// 	StyleName:         "TableStyleMedium2",
		// 	ShowFirstColumn:   true,
		// 	ShowLastColumn:    false,
		// 	ShowRowStripes:    &showStripes,
		// 	ShowColumnStripes: false,
		// })
		// if err != nil {
		// 	log.Fatal(err)
		// }

		for _, mR := range monthRange.Rows {
			s := fmt.Sprintf("%s%d", monthSummary[fields[0]].Column, mR)
			e := fmt.Sprintf("%s%d", monthSummary[fields[len(fields)-1]].Column, mR)
			c.SetStyle(sheet.Name, s, e, styles.monthRow)
			s = fmt.Sprintf("%s%d", monthSummary["Date"].Column, mR)
			e = fmt.Sprintf("%s%d", monthSummary["Date"].Column, mR)
			c.SetStyle(sheet.Name, s, e, styles.monthRowDate)
			s = fmt.Sprintf("%s%d", monthSummary["PrevUsd"].Column, mR)
			e = fmt.Sprintf("%s%d", monthSummary["CheckUsd"].Column, mR)
			c.SetStyle(sheet.Name, s, e, styles.monthRow2)
			s = fmt.Sprintf("%s%d", monthSummary["BegUnits"].Column, mR)
			e = fmt.Sprintf("%s%d", monthSummary["Check"].Column, mR)
			c.SetStyle(sheet.Name, s, e, styles.monthRow5)
		}

		for _, yR := range yearRange.Rows {
			s := fmt.Sprintf("%s%d", yearSummary[fields[0]].Column, yR)
			e := fmt.Sprintf("%s%d", yearSummary[fields[len(fields)-1]].Column, yR)
			c.SetStyle(sheet.Name, s, e, styles.yearRow)
			s = fmt.Sprintf("%s%d", yearSummary["Date"].Column, yR)
			e = fmt.Sprintf("%s%d", yearSummary["Date"].Column, yR)
			c.SetStyle(sheet.Name, s, e, styles.yearRowDate)
			s = fmt.Sprintf("%s%d", yearSummary["PrevUsd"].Column, yR)
			e = fmt.Sprintf("%s%d", yearSummary["CheckUsd"].Column, yR)
			c.SetStyle(sheet.Name, s, e, styles.yearRow2)
			s = fmt.Sprintf("%s%d", yearSummary["BegUnits"].Column, yR)
			e = fmt.Sprintf("%s%d", yearSummary["Check"].Column, yR)
			c.SetStyle(sheet.Name, s, e, styles.yearRow5)
		}
	}

	rowRange := Rng{}
	var cnt = 1
	var iRow = 5
	c.SetCell("Summary", iRow, rowRange, &Field{1, "A", 12, "string", "", styles.mainHeader}, "Ignored Addresses:")
	for k, _ := range c.Ignored {
		c.SetCell("Summary", iRow+cnt, rowRange, &Field{1, "B", 12, "string", "", styles.address1}, k)
		c.SetCell("Summary", iRow+cnt, rowRange, &Field{1, "G", 12, "string", "", styles.address1}, c.Opts.Names[common.HexToAddress(k)].Name)
		cnt++
	}

	excel.WriteLicenseSheet(c.ExcelFile)
	c.ExcelFile.SetActiveSheet(1)
	c.ExcelFile.SaveAs("Book1.xlsx")
	return c.Name() + "\n\t" + c.reportValue("Excel: ", uint64(c.Line))
}

func (c *Excel) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *Excel) Sort(array []*mytypes.RawReconciliation) {
	sort.Slice(array, func(i, j int) bool {
		item1 := array[i]
		item2 := array[j]
		if item1.Date == item2.Date {
			if item1.TransactionIndex == item2.TransactionIndex {
				return item1.LogIndex < item2.LogIndex
			}
			return item1.TransactionIndex < item2.TransactionIndex
		}
		return item1.Date.Before(&item2.Date)
	})
}

func (c *Excel) reportValue(msg string, v uint64) string {
	return fmt.Sprintf("%s%d", msg, v)
}

func (c *Excel) ConvertToSheets() []AssetSheet {
	sheets := make([]AssetSheet, 0, len(c.Assets))
	for _, asset := range c.Assets {
		sheetName := asset[0].AssetAddress.String()[:8]
		if !strings.HasPrefix(asset[0].AssetSymbol, "0x") {
			sym := strings.ReplaceAll(asset[0].AssetSymbol, " ", "_")
			if len(sym) > 10 {
				sym = sym[:10]
			}
			sheetName = sym + "_" + sheetName
		} else {
			sheetName = asset[0].AssetAddress.String()[:12]
		}
		sheetName += "_" + fmt.Sprintf("%d", len(asset))
		addr := asset[0].AssetAddress
		s := AssetSheet{
			Name:     sheetName + fmt.Sprintf(" (%d)", len(asset)),
			Address:  addr.String(),
			Symbol:   asset[0].AssetSymbol,
			Decimals: int(asset[0].Decimals),
			nRecords: len(asset),
			Records:  asset,
		}
		sheets = append(sheets, s)
	}

	sort.Slice(sheets, func(i, j int) bool {
		if sheets[i].nRecords != sheets[j].nRecords {
			return sheets[i].nRecords > sheets[j].nRecords
		}
		if sheets[i].Address != sheets[j].Address {
			return sheets[i].Address < sheets[j].Address
		}
		return sheets[i].Name < sheets[j].Name
	})

	cleanup := func(s string) string {
		for _, ch := range s {
			if !unicode.IsLetter(ch) && !unicode.IsNumber(ch) {
				s = strings.ReplaceAll(s, string(ch), "")
			}
		}
		return s
	}

	for i := 0; i < len(sheets); i++ {
		records := sheets[i].Records[0]
		if !strings.HasPrefix(records.AssetSymbol, "0x") {
			sheets[i].RangeName = "t_" + cleanup(records.AssetSymbol) + "_" + records.AssetAddress.String()[:8]
		} else {
			sheets[i].RangeName = "t_" + records.AssetAddress.String()[:12]
		}
		// fmt.Println(sheets[i].RangeName)
	}

	return sheets
}

func (c *Excel) SetStyle(sheetName, topLeft, bottomRight string, styleId int) {
	err := c.ExcelFile.SetCellStyle(sheetName, topLeft, bottomRight, styleId)
	if err != nil {
		log.Fatal(fmt.Errorf("error SetStyle::Setregular(%s, %s, %s, %d) %w", sheetName, topLeft, bottomRight, styleId, err))
	}
}

func (c *Excel) HeaderCell(sheetName, c1, c2, c3, c4, t, v string) (err error) {
	if err = c.ExcelFile.MergeCell(sheetName, c1, c2); err == nil {
		if err = c.ExcelFile.MergeCell(sheetName, c3, c4); err == nil {
			if err = c.ExcelFile.SetCellValue(sheetName, c1, t); err == nil {
				return c.ExcelFile.SetCellValue(sheetName, c3, v)
			}
		}
	}
	return
}

func (c *Excel) SetHeader(sheet *AssetSheet, styles *Styles, lastCol string) {
	if err := c.HeaderCell(sheet.Name, "A1", "C1", "D1", "H1", "Asset Address:", sheet.Address); err != nil {
		log.Fatal(fmt.Errorf("line %d SetHeader::HeaderCell(\"Address\", %s, %s) %w", c.Line, sheet.Name, sheet.Address, err))
	}

	name := c.Opts.Names[common.HexToAddress(sheet.Address)].Name
	if len(name) == 0 {
		name = "Unnamed"
	}
	if err := c.HeaderCell(sheet.Name, "A2", "C2", "D2", "H2", "Asset Name:", name); err != nil {
		log.Fatal(fmt.Errorf("line %d SetHeader::HeaderCell(\"Name\", %s, %s) %w", c.Line, sheet.Name, name, err))
	}
	if err := c.HeaderCell(sheet.Name, "A3", "C3", "D3", "H3", "Asset Symbol:", sheet.Symbol); err != nil {
		log.Fatal(fmt.Errorf("line %d SetHeader::HeaderCell(\"Symbol\", %s, %s) %w", c.Line, sheet.Name, sheet.Symbol, err))
	}
	if err := c.HeaderCell(sheet.Name, "A4", "C4", "D4", "H4", "Decimals:", fmt.Sprintf("%d", sheet.Decimals)); err != nil {
		log.Fatal(fmt.Errorf("line %d SetHeader::HeaderCell(\"Decimals\", %s, %d) %w", c.Line, sheet.Name, sheet.Decimals, err))
	}

	link, tooltip := "https://etherscan.io/address/"+sheet.Address, "Open in Explorer"
	if err := c.ExcelFile.SetCellHyperLink(sheet.Name, "D1", link, "External", excelize.HyperlinkOpts{
		Tooltip: &tooltip,
	}); err != nil {
		log.Fatal(fmt.Errorf("line %d SetHeader::SetCellHyperLink: %w", c.Line, err))
	}

	c.SetStyle(sheet.Name, "D1", "D1", styles.link)
	c.SetStyle(sheet.Name, fmt.Sprintf("A%d", headerRow), fmt.Sprintf((lastCol+"%d"), headerRow), styles.tableHeader)
	c.SetStyle(sheet.Name, fmt.Sprintf("A%d", 1), fmt.Sprintf("F%d", 4), styles.mainHeader)
}

func (c *Excel) GetStyles() (styles Styles, err error) {
	if c.ExcelFile == nil {
		return
	}

	if styles.link, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family:    "Andale Mono",
			Color:     "#0000FF",
			Underline: "1",
		},
	}); err != nil {
		return
	}

	if styles.regular, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
		},
	}); err != nil {
		return
	}

	if styles.integer, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FF0000",
		},
	}); err != nil {
		return
	}

	acct2Fmt := "_(* #,##0.00_);_(* (#,##0.00);_(* \"-\"??_);_(@_)"
	if styles.accounting2, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#AA00AA",
		},
		CustomNumFmt: &acct2Fmt,
	}); err != nil {
		return
	}

	acct5Fmt := "_(* #,##0.00000_);_(* (#,##0.000000);_(* \"-\"??_);_(@_)"
	if styles.accounting5, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#4444ff",
		},
		CustomNumFmt: &acct5Fmt,
	}); err != nil {
		return
	}

	priceFmt := "#,##0.00_);(#,##0.00)"
	if styles.price, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#AA00AA",
		},
		CustomNumFmt: &priceFmt,
	}); err != nil {
		return
	}

	year := "yyyy"
	if styles.dateYear, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{

			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &year,
	}); err != nil {
		return
	}

	month := "yyyy-mm"
	if styles.dateMonth, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{

			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &month,
	}); err != nil {
		return
	}

	date := "mm/dd/yyyy hh:mm:ss"
	if styles.date, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &date,
	}); err != nil {
		return
	}

	if styles.boolean, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FF0000",
			Bold:   true,
		},
	}); err != nil {
		return
	}

	if styles.address1, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
		},
	}); err != nil {
		return
	}

	if styles.address2, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FFFFFF",
		},
		// Alignment: &excelize.Alignment{
		// 	Horizontal: "center",
		// },
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#33503C"},
		},
	}); err != nil {
		return
	}

	if styles.address3, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FFFFFF",
		},
		// Alignment: &excelize.Alignment{
		// 	Horizontal: "center",
		// },
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#dd6677"},
		},
	}); err != nil {
		return
	}

	if styles.bigInteger, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FFFF00",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#000000"},
		},
		NumFmt: 1,
	}); err != nil {
		return
	}

	if styles.zero, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#888888",
			Italic: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	}); err != nil {
		return
	}

	if styles.mainHeader, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Size:   12,
		},
	}); err != nil {
		return
	}

	if styles.tableHeader, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Family: "Andale Mono",
			Size:   12,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Style: 1, Color: "000000"},
		},
	}); err != nil {
		return
	}

	green := excelize.Fill{
		Type:    "pattern",
		Pattern: 1,
		Color:   []string{"#cbe0b8"},
	}
	yellow := excelize.Fill{
		Type:    "pattern",
		Pattern: 1,
		Color:   []string{"#fbe7a2"},
	}

	if styles.monthRow, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		Fill: green,
	}); err != nil {
		return
	}

	if styles.monthRowDate, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		CustomNumFmt: &month,
		Fill:         green,
	}); err != nil {
		return
	}

	if styles.monthRow2, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		CustomNumFmt: &acct2Fmt,
		Fill:         green,
	}); err != nil {
		return
	}

	if styles.monthRow5, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		CustomNumFmt: &acct5Fmt,
		Fill:         green,
	}); err != nil {
		return
	}

	if styles.yearRow, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		Fill: yellow,
	}); err != nil {
		return
	}

	if styles.yearRowDate, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &year,
		Fill:         yellow,
	}); err != nil {
		return
	}

	if styles.yearRow2, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		CustomNumFmt: &acct2Fmt,
		Fill:         yellow,
	}); err != nil {
		return
	}

	if styles.yearRow5, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#000000",
		},
		CustomNumFmt: &acct5Fmt,
		Fill:         yellow,
	}); err != nil {
		return
	}

	return
}

func (c *Excel) SetCell(sheetName string, row int, sumRange Rng, field *Field, val interface{}) string {
	var err error
	cell := fmt.Sprintf("%s%d", field.Column, row)
	switch field.Format {
	case "int":
		err = c.ExcelFile.SetCellInt(sheetName, cell, val.(int))
	case "formula":
		f := field.Formula
		f = strings.Replace(f, "{R}", strconv.Itoa(row), -1)
		f = strings.Replace(f, "{R-1}", strconv.Itoa(row-1), -1)
		f = strings.Replace(f, "{A}", strconv.Itoa(sumRange.A+1), -1)
		f = strings.Replace(f, "{B}", strconv.Itoa(sumRange.B-1), -1)
		err = c.ExcelFile.SetCellFormula(sheetName, cell, f)
	case "float2":
		err = c.ExcelFile.SetCellFloat(sheetName, cell, val.(float64), 2, 64)
	case "float5":
		err = c.ExcelFile.SetCellFloat(sheetName, cell, val.(float64), 5, 64)
	case "bool":
		err = c.ExcelFile.SetCellBool(sheetName, cell, val.(bool))
	case "big":
		v := val.(string)
		var x big.Float
		x.SetString(v)
		f, _ := x.Float64()
		// log.Fatal(v, f)
		err = c.ExcelFile.SetCellFloat(sheetName, cell, f, 18, 64)
	case "date":
		v := val.(mytypes.DateTime)
		tt := time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond(), v.Location())
		err = c.ExcelFile.SetCellValue(sheetName, cell, tt)
	case "address":
		a := val.(mytypes.Address)
		err = c.ExcelFile.SetCellStr(sheetName, cell, a.String())
	case "hash":
		a := val.(common.Hash)
		err = c.ExcelFile.SetCellStr(sheetName, cell, a.String())
	case "string":
		var s string
		if len(field.Formula) > 0 {
			s = strings.Replace(field.Formula, "{R}", strconv.Itoa(row), -1)
		} else {
			s = val.(string)
		}
		err = c.ExcelFile.SetCellStr(sheetName, cell, s)
	default:
		err = c.ExcelFile.SetCellValue(sheetName, cell, val)
	}

	if err != nil {
		log.Fatal(err)
	}

	return cell
}

func ToUnits(amount *big.Float, decimals uint64) *big.Float {
	if amount == utils.Zero() || decimals == 0 {
		return utils.Zero()
	}
	ten := new(big.Float)
	ten = ten.SetFloat64(10)
	divisor := utils.Pow(ten, decimals)
	if divisor == nil || divisor == utils.Zero() {
		return nil
	}
	return utils.Zero().Quo(amount, divisor)
}
