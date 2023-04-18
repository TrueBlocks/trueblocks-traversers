package accounting

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

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

func (c *Excel) SetStyle(sheetName, topLeft, bottomRight string, styleId int) {
	err := c.ExcelFile.SetCellStyle(sheetName, topLeft, bottomRight, styleId)
	if err != nil {
		log.Fatal(fmt.Errorf("error SetStyle::Setregular(%s, %s, %s, %d) %w", sheetName, topLeft, bottomRight, styleId, err))
	}
}
