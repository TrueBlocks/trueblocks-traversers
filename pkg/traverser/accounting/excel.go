package accounting

import (
	"accounting/pkg/excel"
	"accounting/pkg/mytypes"
	"accounting/pkg/traverser"
	"accounting/pkg/utils"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
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

type AssetSheet struct {
	Name     string
	Address  string
	Symbol   string
	Decimals int
	nRecords int
	Records  []*mytypes.RawReconciliation
}

type Field struct {
	Column string
	Wid    float64
	Format string
	Style  int
}

const headerRow = 6

func (c *Excel) Result() string {
	cellStyle, intStyle, float2Style, float5Style, yearStyle, monthStyle, dateStyle, boolStyle, addrStyle1, addrStyle2, addrStyle3, bigStyle, zeroStyle, err := c.GetStyles()
	if err != nil {
		panic(err)
	}

	// FIELD_ORDER
	var fields = []string{
		"Type",
		"Bn",
		"TxId",
		"Year",
		"Month",
		"Date",
		"PrevUsd",
		"ChangeUsd",
		"BegUsd",
		"NetUsd",
		"EndUsd",
		"Spot",
		"Source",
		"BegUnits",
		"NetUnits",
		"EndUnits",
		"BegBal",
		"AmountNet",
		"EndBal",
		"Check",
		"Message",
		"Sender",
		"Recipient",
		"AccountedFor",
	}

	var fieldMap = map[string]*Field{
		"Type":         {"A", 5, "string", cellStyle},
		"Bn":           {"B", 12, "int", intStyle},
		"TxId":         {"C", 8, "int", intStyle},
		"Year":         {"D", 12, "date", yearStyle},
		"Month":        {"E", 12, "date", monthStyle},
		"Date":         {"F", 0, "date", dateStyle},
		"PrevUsd":      {"G", 18, "float2", float2Style},
		"ChangeUsd":    {"H", 18, "float2", float2Style},
		"BegUsd":       {"I", 18, "float2", float2Style},
		"NetUsd":       {"J", 18, "float2", float2Style},
		"EndUsd":       {"K", 18, "float2", float2Style},
		"Spot":         {"L", 18, "float2", float2Style},
		"Source":       {"M", 10, "string", cellStyle},
		"BegUnits":     {"N", 18, "float5", float5Style},
		"NetUnits":     {"O", 18, "float5", float5Style},
		"EndUnits":     {"P", 18, "float5", float5Style},
		"BegBal":       {"Q", 0, "big", bigStyle},
		"AmountNet":    {"R", 0, "big", bigStyle},
		"EndBal":       {"S", 0, "big", bigStyle},
		"Check":        {"T", 15, "string", boolStyle},
		"Message":      {"U", 25, "string", cellStyle},
		"Sender":       {"V", 52, "address", addrStyle1},
		"Recipient":    {"W", 52, "address", addrStyle1},
		"AccountedFor": {"X", 0, "address", addrStyle1},
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

			var x big.Float
			x.SetString(r.BegBal)
			begUnits, _ := ToUnits(&x, r.Decimals).Float64()
			x.SetString(r.AmountNet)
			amtUnits, _ := ToUnits(&x, r.Decimals).Float64()
			x.SetString(r.EndBal)
			endUnits, _ := ToUnits(&x, r.Decimals).Float64()
			check := ""
			if !r.Reconciled {
				check = "X"
			}

			row := headerRow + i + 1
			// FIELD_ORDER
			c.SetCell(sheet.Name, row, fieldMap["Type"], "Tx")
			c.SetCell(sheet.Name, row, fieldMap["Bn"], int(r.BlockNumber))
			c.SetCell(sheet.Name, row, fieldMap["TxId"], int(r.TransactionIndex))
			c.SetCell(sheet.Name, row, fieldMap["Year"], r.Date.EndOfYear())
			c.SetCell(sheet.Name, row, fieldMap["Month"], r.Date.EndOfMonth())
			c.SetCell(sheet.Name, row, fieldMap["Date"], r.Date)
			c.SetCell(sheet.Name, row, fieldMap["PrevUsd"], prevEnd)
			c.SetCell(sheet.Name, row, fieldMap["ChangeUsd"], (r.SpotPrice*begUnits)-prevEnd)
			c.SetCell(sheet.Name, row, fieldMap["BegUsd"], r.SpotPrice*begUnits)
			c.SetCell(sheet.Name, row, fieldMap["NetUsd"], r.SpotPrice*amtUnits)
			c.SetCell(sheet.Name, row, fieldMap["EndUsd"], r.SpotPrice*endUnits)
			c.SetCell(sheet.Name, row, fieldMap["Spot"], r.SpotPrice)
			c.SetCell(sheet.Name, row, fieldMap["Source"], r.PriceSource)
			c.SetCell(sheet.Name, row, fieldMap["BegUnits"], begUnits)
			c.SetCell(sheet.Name, row, fieldMap["NetUnits"], amtUnits)
			c.SetCell(sheet.Name, row, fieldMap["EndUnits"], endUnits)
			c.SetCell(sheet.Name, row, fieldMap["Check"], check)
			c.SetCell(sheet.Name, row, fieldMap["Message"], sig)
			c.SetCell(sheet.Name, row, fieldMap["BegBal"], r.BegBal)
			c.SetCell(sheet.Name, row, fieldMap["AmountNet"], r.AmountNet)
			c.SetCell(sheet.Name, row, fieldMap["EndBal"], r.EndBal)
			c.SetCell(sheet.Name, row, fieldMap["Sender"], r.Sender)
			c.SetCell(sheet.Name, row, fieldMap["Recipient"], r.Recipient)
			c.SetCell(sheet.Name, row, fieldMap["AccountedFor"], r.AccountedFor)

			// both or neither can be true...
			senderCell := fmt.Sprintf("%s%d", fieldMap["Sender"].Column, row)
			if r.Sender.IsZero() {
				c.SetStyle(sheet.Name, senderCell, senderCell, zeroStyle)
			} else {
				style := addrStyle1
				if c.Opts.Names[common.HexToAddress(r.Sender.String())].IsCustom {
					style = addrStyle3
				}
				if r.Sender == r.AccountedFor {
					style = addrStyle2
				}
				c.SetStyle(sheet.Name, senderCell, senderCell, style)
			}

			recipCell := fmt.Sprintf("%s%d", fieldMap["Recipient"].Column, row)
			if r.Recipient.IsZero() {
				c.SetStyle(sheet.Name, recipCell, recipCell, zeroStyle)
			} else {
				style := addrStyle1
				if c.Opts.Names[common.HexToAddress(r.Recipient.String())].IsCustom {
					style = addrStyle3
				}
				if r.Recipient == r.AccountedFor {
					style = addrStyle2
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

		// err = c.ExcelFile.AddTable(sheet.Name, "A1:"+last, &excelize.TableOptions{
		// 	Name:      sheet.Name,
		// 	StyleName: "TableStyleMedium2",
		// })
		// if err != nil {
		// 	log.Fatal(err)
		// }
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
		// if len(asset) < 10 || asset[0].AssetSymbol != "WEI" {
		// 	continue
		// }
		sheetName := asset[0].AssetSymbol
		if len(sheetName) == 0 {
			sheetName = asset[0].AssetAddress.String()
		}
		if len(sheetName) > 10 {
			sheetName = sheetName[:10]
		}
		sheets = append(sheets, AssetSheet{
			Name:     sheetName + fmt.Sprintf(" (%d)", len(asset)),
			Address:  asset[0].AssetAddress.String(),
			Symbol:   asset[0].AssetSymbol,
			Decimals: int(asset[0].Decimals),
			nRecords: len(asset),
			Records:  asset,
		})
	}
	sort.Slice(sheets, func(i, j int) bool {
		if sheets[i].nRecords != sheets[j].nRecords {
			return sheets[i].nRecords > sheets[j].nRecords
		}
		return sheets[i].Name < sheets[j].Name
	})

	return sheets
}

func (c *Excel) SetStyle(sheetName, topLeft, bottomRight string, styleId int) {
	err := c.ExcelFile.SetCellStyle(sheetName, topLeft, bottomRight, styleId)
	if err != nil {
		log.Fatal(fmt.Errorf("error SetStyle::SetCellStyle(%s, %s, %s, %d) %w", sheetName, topLeft, bottomRight, styleId, err))
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
		log.Fatal(err) // fmt.Errorf("line %d SetHeader::HeaderCell(\"Address\", %s, %s) %w", c.Line, sheet.Name, sheet.Address, err))
	}
	name := c.Opts.Names[common.HexToAddress(sheet.Address)].Name
	if len(name) == 0 {
		name = "Unnamed"
	}
	if err := c.HeaderCell(sheet.Name, "A2", "C2", "D2", "H2", "Asset Name:", name); err != nil {
		log.Fatal(err) // fmt.Errorf("line %d SetHeader::HeaderCell(\"Name\", %s, %s) %w", c.Line, sheet.Name, name, err))
	}
	if err := c.HeaderCell(sheet.Name, "A3", "C3", "D3", "H3", "Asset Symbol:", sheet.Symbol); err != nil {
		log.Fatal(err) // fmt.Errorf("line %d SetHeader::HeaderCell(\"Symbol\", %s, %s) %w", c.Line, sheet.Name, sheet.Symbol, err))
	}
	if err := c.HeaderCell(sheet.Name, "A4", "C4", "D4", "H4", "Decimals:", fmt.Sprintf("%d", sheet.Decimals)); err != nil {
		log.Fatal(err) // fmt.Errorf("line %d SetHeader::HeaderCell(\"Decimals\", %s, %d) %w", c.Line, sheet.Name, sheet.Decimals, err))
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
		c.SetStyle(sheet.Name, fmt.Sprintf("A%d", headerRow), fmt.Sprintf("X%d", headerRow), tableHeader)
	}
}

func (c *Excel) GetStyles() (cellStyle, intStyle, float2Style, float5Style, yearStyle, monthStyle, dateStyle, boolStyle, addrStyle1, addrStyle2, addrStyle3, bigStyle, zeroStyle int, err error) {
	if cellStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
		},
	}); err != nil {
		return
	}

	if intStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FF0000",
		},
	}); err != nil {
		return
	}

	bh1 := "#,##0.00"
	if float2Style, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#AA00AA",
		},
		CustomNumFmt: &bh1,
	}); err != nil {
		return
	}

	bh := "#,##0.00000"
	if float5Style, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#4444ff",
		},
		CustomNumFmt: &bh,
	}); err != nil {
		return
	}

	year := "YYYY"
	if yearStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{

			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &year,
	}); err != nil {
		return
	}

	month := "YYYY-mm"
	if monthStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{

			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &month,
	}); err != nil {
		return
	}

	date := "mm/dd/yyyy hh:mm:ss"
	if dateStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#0000FF",
		},
		CustomNumFmt: &date,
	}); err != nil {
		return
	}

	if boolStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
			Color:  "#FF0000",
			Bold:   true,
		},
	}); err != nil {
		return
	}

	if addrStyle1, err = c.ExcelFile.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Andale Mono",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	}); err != nil {
		return
	}

	if addrStyle2, err = c.ExcelFile.NewStyle(&excelize.Style{
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

	if addrStyle3, err = c.ExcelFile.NewStyle(&excelize.Style{
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

	if bigStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
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

	if zeroStyle, err = c.ExcelFile.NewStyle(&excelize.Style{
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
