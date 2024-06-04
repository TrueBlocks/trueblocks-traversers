package accounting

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/excel"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/xuri/excelize/v2"
)

type Excel struct {
	Opts      traverser.Options
	ExcelFile *excelize.File
	Line      int
	Assets    map[string][]*mytypes.RawReconciliation
	showOnce  map[string]bool
}

// --------------------------------
func (c *Excel) Traverse(r *mytypes.RawReconciliation) {
	if c.ExcelFile == nil {
		c.ExcelFile = excel.NewWorkbook("Summary", []string{"This is the summary text"})
		c.Assets = make(map[string][]*mytypes.RawReconciliation)
		c.showOnce = make(map[string]bool)
	}

	passedAddrFilter := c.Opts.AddrFilters[r.AssetAddress]
	if len(c.Opts.AddrFilters) > 0 && !passedAddrFilter {
		if !c.showOnce[r.AssetAddress.String()] {
			fmt.Println("Skipping asset", r.AssetAddress.String())
		}
		c.showOnce[r.AssetAddress.String()] = true
		return
	}

	l := 0 // len(c.Opts.DateFilters)
	if l > 0 {
		firstDate := mytypes.NewDateTime2(2015, 7, 30, 23, 59, 59)
		lastDate := c.Opts.DateFilters[0]
		if l > 1 {
			firstDate = c.Opts.DateFilters[0]
			lastDate = c.Opts.DateFilters[1]
		}
		if r.Date.Before(&firstDate) || r.Date.After(&lastDate) {
			month := r.Date.String()[:7]
			if !c.showOnce[month] {
				fmt.Println("Skipping month", month)
			}
			c.showOnce[month] = true
			return
		}
	}

	c.Line += 1
	c.Assets[c.GetKey(r)] = append(c.Assets[c.GetKey(r)], r)
}

func (c *Excel) GetKey(r *mytypes.RawReconciliation) string {
	return fmt.Sprintf("%s-%s-%d", r.AssetAddress.String(), r.AssetSymbol, r.Decimals)
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

type Field struct {
	Order   int
	Column  string
	Wid     float64
	Format  string
	Formula string
	Style   int
}

const headerRow = 6

type CellRange struct {
	pDate   mytypes.DateTime
	A       int
	B       int
	Rows    []int
	CurRows []int
}

func (c *Excel) Result() string {
	styles, err := c.GetStyles()
	if err != nil {
		panic(err)
	}

	var fieldMap = map[string]*Field{
		"Type":            {1, "A", 4, "string", "", styles.regular},
		"Bn":              {2, "B", 11, "int", "", styles.integer},
		"TxId":            {3, "C", 6, "int", "", styles.integer},
		"LogId":           {4, "D", 6, "int", "", styles.integer},
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
		"Source":          {16, "P", 12, "string", "", styles.regular},
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
		"Check":           {27, "AA", 15, "formula", "=ROUND(Q{R}+R{R}-S{R}-T{R}-U{R},5)", styles.accounting5},
		"Message":         {30, "AB", 15, "string", "", styles.regular},
		"ReconType":       {31, "AC", 0, "string", "", styles.regular},
		"Sender":          {32, "AD", 25, "address", "", styles.address1},
		"Recipient":       {33, "AE", 25, "address", "", styles.address1},
		"AccountedFor":    {34, "AF", 25, "address", "", styles.address1},
		"TransactionHash": {35, "AG", 80, "hash", "", styles.link},
	}
	lastCol := "AG"

	var monthly = map[string]*Field{
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

	var annually = map[string]*Field{
		"Date":      {1, "G", 25, "date", "", styles.date},
		"PrevUsd":   {2, "H", 15, "formula", "=H{L0}", styles.yearRow2},
		"ChangeUsd": {3, "I", 15, "formula", "={L}", styles.yearRow2},
		"BegUsd":    {4, "J", 15, "formula", "=J{L0}", styles.yearRow2},
		"InUsd":     {5, "K", 15, "formula", "={L}", styles.yearRow2},
		"OutUsd":    {6, "L", 15, "formula", "={L}", styles.yearRow2},
		"GasUsd":    {7, "M", 15, "formula", "={L}", styles.yearRow2},
		"EndUsd":    {8, "N", 15, "formula", "=N{LN-1}", styles.yearRow2},
		"CheckUsd":  {9, "O", 15, "formula", "=H{R}+I{R}+K{R}-L{R}-M{R}-N{R}", styles.price},
		"BegUnits":  {10, "Q", 18, "formula", "=Q{L0}", styles.yearRow5},
		"InUnits":   {11, "R", 18, "formula", "={L}", styles.yearRow5},
		"OutUnits":  {12, "S", 18, "formula", "={L}", styles.yearRow5},
		"GasUnits":  {13, "T", 18, "formula", "={L}", styles.yearRow5},
		"EndUnits":  {14, "U", 18, "formula", "=U{LN-1}", styles.yearRow5},
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

	sheets := c.assetsToSheets()
	for _, sheet := range sheets {
		c.ExcelFile.NewSheet(sheet.Name)
		c.SetHeader(&sheet, &styles, lastCol)
		for _, field := range fields {
			c.ExcelFile.SetColWidth(sheet.Name, fieldMap[field].Column, fieldMap[field].Column, fieldMap[field].Wid)
		}

		bottomRow := headerRow
		verbose := c.Opts.Verbose != 0

		c.ExcelFile.SetSheetRow(sheet.Name, fmt.Sprintf("A%d", headerRow), &fields)

		rowRange := CellRange{}
		monthRange := CellRange{
			A: headerRow,
			B: headerRow,
		}
		yearRange := CellRange{
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

			if sheet.monthSwitches(txIndex) {
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
					c.SetCell(sheet.Name, curRow, monthRange, monthly["Date"], monthRange.pDate)
					for k, field := range monthly {
						if k != "Date" {
							c.SetCell(sheet.Name, curRow, monthRange, field, "")
						}
					}

					monthRange.A = curRow
					lastRowType = "Mo"
					if sheet.yearSwitches(txIndex) {
						curRow++
						yearRange.Rows = append(yearRange.Rows, curRow)
						yearRange.CurRows = append(yearRange.CurRows, curRow)
						yearRange.CurRows = monthRange.CurRows
						msg := "Yr"
						if verbose {
							msg = fmt.Sprintf("%s-[%03d] [%s]", "Yr", curRow, toStr(yearRange.CurRows))
						}
						c.SetCell(sheet.Name, curRow, yearRange, fieldMap["Type"], msg)
						c.SetCell(sheet.Name, curRow, yearRange, annually["Date"], yearRange.pDate)
						for k, field := range annually {
							if k != "Date" {
								c.SetCell(sheet.Name, curRow, yearRange, field, "")
							}
						}
						monthRange.CurRows = []int{}
						monthRange.A = curRow
						lastRowType = "Yr"
					}
				}
			}

			if true {
				curRow++
				lastRowType = "Tx"
				msg := "Tx"
				if verbose {
					msg = fmt.Sprintf("%s-%d", "Tx", txIndex)
				}
				c.SetCell(sheet.Name, curRow, monthRange, fieldMap["Type"], msg)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Bn"], int(r.BlockNumber))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["TxId"], int(r.TransactionIndex))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["LogId"], int(r.LogIndex))
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Year"], r.Date.EndOfYear())
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Month"], r.Date.EndOfMonth())
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["Date"], r.Date)
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["PrevUsd"], "")
				c.SetCell(sheet.Name, curRow, rowRange, fieldMap["ChangeUsd"], "")
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
				c.setLink(sheet.Name, fieldMap["TransactionHash"].Cell(curRow), "https://etherscan.io/tx/"+r.TransactionHash.String(), "View on Etherscan")

				// both or neither can be true...
				senderCell := fmt.Sprintf("%s%d", fieldMap["Sender"].Column, curRow)
				if r.Sender.IsZero() {
					c.setStyle(sheet.Name, senderCell, senderCell, styles.zero)
				} else {
					style := styles.address1
					if c.Opts.Names[base.HexToAddress(r.Sender.String())].IsCustom {
						style = styles.address3
					}
					if r.Sender == r.AccountedFor {
						style = styles.address2
					}
					c.setStyle(sheet.Name, senderCell, senderCell, style)
				}

				recipCell := fmt.Sprintf("%s%d", fieldMap["Recipient"].Column, curRow)
				if r.Recipient.IsZero() {
					c.setStyle(sheet.Name, recipCell, recipCell, styles.zero)
				} else {
					style := styles.address1
					if c.Opts.Names[base.HexToAddress(r.Recipient.String())].IsCustom {
						style = styles.address3
					}
					if r.Recipient == r.AccountedFor {
						style = styles.address2
					}
					c.setStyle(sheet.Name, recipCell, recipCell, style)
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
			c.SetCell(sheet.Name, curRow, monthRange, monthly["Date"], monthRange.pDate)
			for k, field := range monthly {
				if k != "Date" {
					c.SetCell(sheet.Name, curRow, monthRange, field, "")
				}
			}
			monthRange.A = curRow
			lastRowType = "Mo"
		}

		if lastRowType == "Mo" || lastRowType == "" {
			curRow++
			yearRange.Rows = append(yearRange.Rows, curRow)
			yearRange.CurRows = append(yearRange.CurRows, curRow)
			yearRange.CurRows = monthRange.CurRows
			msg := "Yr"
			if verbose {
				msg = fmt.Sprintf("%s-[%03d] [%s]", "Yr", curRow, toStr(yearRange.CurRows))
			}
			c.SetCell(sheet.Name, curRow, yearRange, fieldMap["Type"], msg)
			c.SetCell(sheet.Name, curRow, yearRange, annually["Date"], yearRange.pDate)
			for k, field := range annually {
				if k != "Date" {
					c.SetCell(sheet.Name, curRow, yearRange, field, "")
				}
			}
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
			c.setStyle(sheet.Name, cell1, cell2, style)
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
			s := fmt.Sprintf("%s%d", fieldMap[fields[0]].Column, mR)
			e := fmt.Sprintf("%s%d", fieldMap[fields[len(fields)-1]].Column, mR)
			c.setStyle(sheet.Name, s, e, styles.monthRow)
			s = fmt.Sprintf("%s%d", monthly["Date"].Column, mR)
			e = fmt.Sprintf("%s%d", monthly["Date"].Column, mR)
			c.setStyle(sheet.Name, s, e, styles.monthRowDate)
			s = fmt.Sprintf("%s%d", monthly["PrevUsd"].Column, mR)
			e = fmt.Sprintf("%s%d", monthly["CheckUsd"].Column, mR)
			c.setStyle(sheet.Name, s, e, styles.monthRow2)
			s = fmt.Sprintf("%s%d", monthly["BegUnits"].Column, mR)
			e = fmt.Sprintf("%s%d", monthly["Check"].Column, mR)
			c.setStyle(sheet.Name, s, e, styles.monthRow5)
		}

		for _, yR := range yearRange.Rows {
			s := fmt.Sprintf("%s%d", fieldMap["Type"].Column, yR)
			e := fmt.Sprintf("%s%d", fieldMap[fields[len(fields)-1]].Column, yR)
			c.setStyle(sheet.Name, s, e, styles.yearRow)
			s = fmt.Sprintf("%s%d", annually["Date"].Column, yR)
			e = fmt.Sprintf("%s%d", annually["Date"].Column, yR)
			c.setStyle(sheet.Name, s, e, styles.yearRowDate)
			s = fmt.Sprintf("%s%d", annually["PrevUsd"].Column, yR)
			e = fmt.Sprintf("%s%d", annually["CheckUsd"].Column, yR)
			c.setStyle(sheet.Name, s, e, styles.yearRow2)
			s = fmt.Sprintf("%s%d", annually["BegUnits"].Column, yR)
			e = fmt.Sprintf("%s%d", annually["Check"].Column, yR)
			c.setStyle(sheet.Name, s, e, styles.yearRow5)
		}
	}

	// rowRange := CellRange{}
	// var cnt = 1
	// var iRow = 5
	// c.SetCell("Summary", iRow, rowRange, &Field{1, "A", 12, "string", "", styles.mainHeader}, "Ignored Addresses:")
	// for k, _ := range c.Ignored {
	// 	c.SetCell("Summary", iRow+cnt, rowRange, &Field{1, "B", 12, "string", "", styles.address1}, k)
	// 	c.SetCell("Summary", iRow+cnt, rowRange, &Field{1, "G", 12, "string", "", styles.address1}, c.Opts.Names[base.HexToAddress(k)].Name)
	// 	cnt++
	// }

	excel.WriteLicenseSheet(c.ExcelFile)
	c.ExcelFile.SetActiveSheet(1)
	c.ExcelFile.SaveAs("Book1.xlsx")
	return c.Name() + "\n\t" + fmt.Sprintf("%s%d", "Excel:", c.Line)
}

func (c *Excel) assetsToSheets() []AssetSheet {
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

func toStr(rows []int) string {
	str := ""
	for _, row := range rows {
		str += fmt.Sprintf("%d,", row)
	}
	return str
}

func (f *Field) Cell(row int) string {
	return fmt.Sprintf("%s%d", f.Column, row)
}
