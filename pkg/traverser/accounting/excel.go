package accounting

import (
	"fmt"
	"log"
	"math/big"
	"reflect"
	"sort"
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
}

func (c *Excel) Traverse(r *mytypes.RawReconciliation) {
	if c.ExcelFile == nil {
		c.ExcelFile = excel.NewWorkbook("Summary", []string{"This is the summary text"})
		c.Assets = make(map[string][]*mytypes.RawReconciliation)
	}
	c.Line += 1
	c.Assets[c.GetKey(r)] = append(c.Assets[c.GetKey(r)], r)
}

func (c *Excel) GetKey(r *mytypes.RawReconciliation) string {
	return fmt.Sprintf("%s-%s-%d", r.AssetAddress.String(), r.AssetSymbol, r.Decimals)
}

var tableCnt = 1

type AssetSheet struct {
	Name      string
	RangeName string
	Address   string
	Symbol    string
	Decimals  int
	nRecords  int
	Records   []*mytypes.RawReconciliation
}

type Field struct {
	Order  int
	Column string
	Wid    float64
	Format string
	Style  int
}

const headerRow = 6

type Styles struct {
	regular     int
	integer     int
	accounting1 int
	accounting5 int
	price       int
	dateYear    int
	dateMonth   int
	date        int
	boolean     int
	address1    int
	address2    int
	address3    int
	bigInteger  int
	zero        int
	link        int
}

func (c *Excel) Result() string {
	styles, err := c.GetStyles()
	if err != nil {
		panic(err)
	}

	var fieldMap = map[string]*Field{
		"Type":            {1, "A", 6, "string", styles.regular},
		"Bn":              {2, "B", 12, "int", styles.integer},
		"TxId":            {3, "C", 7, "int", styles.integer},
		"LogId":           {4, "D", 7, "int", styles.integer},
		"Year":            {5, "E", 8, "date", styles.dateYear},
		"Month":           {6, "F", 10, "date", styles.dateMonth},
		"Date":            {7, "G", 0, "date", styles.date},
		"PrevUsd":         {8, "H", 15, "float2", styles.accounting1},
		"ChangeUsd":       {9, "I", 15, "float2", styles.accounting1},
		"BegUsd":          {10, "J", 15, "float2", styles.accounting1},
		"InUsd":           {11, "K", 15, "float2", styles.accounting1},
		"OutUsd":          {12, "L", 15, "float2", styles.accounting1},
		"GasUsd":          {13, "M", 15, "float2", styles.accounting1},
		"EndUsd":          {14, "N", 15, "float2", styles.accounting1},
		"Spot":            {15, "O", 15, "float2", styles.price},
		"Source":          {16, "P", 15, "string", styles.regular},
		"BegUnits":        {17, "Q", 18, "float5", styles.accounting5},
		"InUnits":         {18, "R", 18, "float5", styles.accounting5},
		"OutUnits":        {19, "S", 18, "float5", styles.accounting5},
		"GasUnits":        {20, "T", 18, "float5", styles.accounting5},
		"EndUnits":        {21, "U", 18, "float5", styles.accounting5},
		"BegBal":          {22, "V", 0, "big", styles.bigInteger},
		"Inflow":          {23, "W", 0, "big", styles.bigInteger},
		"Outflow":         {24, "X", 0, "big", styles.bigInteger},
		"GasOut":          {25, "Y", 0, "big", styles.bigInteger},
		"EndBal":          {26, "Z", 0, "big", styles.bigInteger},
		"Check":           {27, "AA", 10, "string", styles.boolean},
		"Message":         {28, "AB", 20, "string", styles.regular},
		"ReconType":       {29, "AC", 0, "string", styles.regular},
		"Sender":          {30, "AD", 52, "address", styles.address1},
		"Recipient":       {31, "AE", 52, "address", styles.address1},
		"AccountedFor":    {32, "AF", 0, "address", styles.address1},
		"TransactionHash": {33, "AG", 0, "hash", styles.address1},
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
		c.SetHeader(&sheet)
		for _, field := range fields {
			c.ExcelFile.SetColWidth(sheet.Name, fieldMap[field].Column, fieldMap[field].Column, fieldMap[field].Wid)
		}

		bottomRow := headerRow
		prevEnd := float64(0)

		c.ExcelFile.SetSheetRow(sheet.Name, fmt.Sprintf("A%d", headerRow), &fields)

		for i, r := range sheet.Records {
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

			check := ""
			if !r.Reconciled {
				check = "X"
			}

			row := headerRow + i + 1
			// FIELD_ORDER
			c.SetCell(sheet.Name, row, fieldMap["Type"], "Tx")
			c.SetCell(sheet.Name, row, fieldMap["Bn"], int(r.BlockNumber))
			c.SetCell(sheet.Name, row, fieldMap["TxId"], int(r.TransactionIndex))
			c.SetCell(sheet.Name, row, fieldMap["LogId"], int(r.LogIndex))
			c.SetCell(sheet.Name, row, fieldMap["Year"], r.Date.EndOfYear())
			c.SetCell(sheet.Name, row, fieldMap["Month"], r.Date.EndOfMonth())
			c.SetCell(sheet.Name, row, fieldMap["Date"], r.Date)
			c.SetCell(sheet.Name, row, fieldMap["PrevUsd"], prevEnd)
			c.SetCell(sheet.Name, row, fieldMap["ChangeUsd"], (r.SpotPrice*begUnits)-prevEnd)
			c.SetCell(sheet.Name, row, fieldMap["BegUsd"], r.SpotPrice*begUnits)
			c.SetCell(sheet.Name, row, fieldMap["InUsd"], r.SpotPrice*inUnits)
			c.SetCell(sheet.Name, row, fieldMap["OutUsd"], r.SpotPrice*outUnitsLessGas)
			c.SetCell(sheet.Name, row, fieldMap["GasUsd"], r.SpotPrice*gasUnitsOut)
			c.SetCell(sheet.Name, row, fieldMap["EndUsd"], r.SpotPrice*endUnits)
			c.SetCell(sheet.Name, row, fieldMap["Spot"], r.SpotPrice)
			c.SetCell(sheet.Name, row, fieldMap["Source"], r.PriceSource)
			c.SetCell(sheet.Name, row, fieldMap["BegUnits"], begUnits)
			c.SetCell(sheet.Name, row, fieldMap["InUnits"], inUnits)
			c.SetCell(sheet.Name, row, fieldMap["OutUnits"], outUnitsLessGas)
			c.SetCell(sheet.Name, row, fieldMap["GasUnits"], gasUnitsOut)
			c.SetCell(sheet.Name, row, fieldMap["EndUnits"], endUnits)
			c.SetCell(sheet.Name, row, fieldMap["BegBal"], r.BegBal)
			c.SetCell(sheet.Name, row, fieldMap["Inflow"], r.TotalIn)
			c.SetCell(sheet.Name, row, fieldMap["Outflow"], r.TotalOutLessGas)
			c.SetCell(sheet.Name, row, fieldMap["GasOut"], r.GasOut)
			c.SetCell(sheet.Name, row, fieldMap["EndBal"], r.EndBal)
			c.SetCell(sheet.Name, row, fieldMap["Check"], check)
			c.SetCell(sheet.Name, row, fieldMap["Message"], sig)
			c.SetCell(sheet.Name, row, fieldMap["ReconType"], r.ReconciliationType)
			c.SetCell(sheet.Name, row, fieldMap["Sender"], r.Sender)
			c.SetCell(sheet.Name, row, fieldMap["Recipient"], r.Recipient)
			c.SetCell(sheet.Name, row, fieldMap["AccountedFor"], r.AccountedFor)
			c.SetCell(sheet.Name, row, fieldMap["TransactionHash"], r.TransactionHash)

			// both or neither can be true...
			senderCell := fmt.Sprintf("%s%d", fieldMap["Sender"].Column, row)
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

			recipCell := fmt.Sprintf("%s%d", fieldMap["Recipient"].Column, row)
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
			// both or neither can be true...

			bottomRow = row
			prevEnd = r.SpotPrice * endUnits
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

		tableCnt++
		showStripes := true
		cellRange := fmt.Sprintf("%s%d:%s%d", fieldMap[fields[0]].Column, headerRow, fieldMap[fields[len(fields)-1]].Column, bottomRow)
		err = c.ExcelFile.AddTable(sheet.Name, cellRange, &excelize.TableOptions{
			// Name:              fmt.Sprintf("t_%d", tableCnt-1), // sheet.RangeName,
			Name:              sheet.RangeName,
			StyleName:         "TableStyleMedium2",
			ShowFirstColumn:   true,
			ShowLastColumn:    false,
			ShowRowStripes:    &showStripes,
			ShowColumnStripes: false,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	excel.WriteLicenseSheet(c.ExcelFile)
	c.ExcelFile.SetActiveSheet(1)
	c.ExcelFile.SaveAs("Book1.xlsx")
	return c.Name() + "\n\t" + c.reportValue("Excel: ", uint64(c.Line))
}

func (c *Excel) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
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

func (c *Excel) SetHeader(sheet *AssetSheet) {
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

	if mainHeader, err := c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Size:   12,
		},
	}); err != nil {
		return
	} else {
		c.SetStyle(sheet.Name, fmt.Sprintf("A%d", 1), fmt.Sprintf("F%d", 4), mainHeader)
	}

	styles, err := c.GetStyles()
	if err != nil {
		panic(err)
	}
	link, tooltip := "https://etherscan.io/address/"+sheet.Address, "Open in Explorer"
	// fmt.Println(link)
	if err := c.ExcelFile.SetCellHyperLink(sheet.Name, "D1", link, "External", excelize.HyperlinkOpts{
		Tooltip: &tooltip,
	}); err != nil {
		log.Fatal(fmt.Errorf("line %d SetHeader::SetCellHyperLink: %w", c.Line, err))
	}
	c.SetStyle(sheet.Name, "D1", "D1", styles.link)

	if tableHeader, err := c.ExcelFile.NewStyle(&excelize.Style{
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
	} else {
		c.SetStyle(sheet.Name, fmt.Sprintf("A%d", headerRow), fmt.Sprintf("AG%d", headerRow), tableHeader)
	}
}

func (c *Excel) GetStyles() (styles Styles, err error) {
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

	bh1 := "_(* #,##0.00_);_(* (#,##0.00);_(* \"-\"??_);_(@_)"
	if styles.accounting1, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#AA00AA",
		},
		CustomNumFmt: &bh1,
	}); err != nil {
		return
	}

	bh := "_(* #,##0.00000_);_(* (#,##0.000000);_(* \"-\"??_);_(@_)"
	if styles.accounting5, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#4444ff",
		},
		CustomNumFmt: &bh,
	}); err != nil {
		return
	}

	bh3 := "#,##0.00_);(#,##0.00)"
	if styles.price, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#AA00AA",
		},
		CustomNumFmt: &bh3,
	}); err != nil {
		return
	}

	year := "YYYY"
	if styles.dateYear, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{

			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &year,
	}); err != nil {
		return
	}

	month := "YYYY-mm"
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
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
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
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
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

	return
}

func (c *Excel) SetCell(sheetName string, row int, field *Field, val interface{}) string {
	var err error
	cell := fmt.Sprintf("%s%d", field.Column, row)
	switch field.Format {
	case "int":
		err = c.ExcelFile.SetCellInt(sheetName, cell, val.(int))
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
		err = c.ExcelFile.SetCellStr(sheetName, cell, val.(string))
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
