package accounting

import (
	"fmt"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/xuri/excelize/v2"
)

func (c *Excel) SetHeader(sheet *AssetSheet, styles *Styles, lastCol string) error {
	if err := c.headerCell(sheet.Name, "A1", "C1", "D1", "H1", "Asset Address:", sheet.Address); err != nil {
		return err
	}

	name := c.Opts.Names[base.HexToAddress(sheet.Address)].Name
	if len(name) == 0 {
		name = "Unnamed"
	}
	if err := c.headerCell(sheet.Name, "A2", "C2", "D2", "H2", "Asset Name:", name); err != nil {
		return err
	}
	if err := c.headerCell(sheet.Name, "A3", "C3", "D3", "H3", "Asset Symbol:", sheet.Symbol); err != nil {
		return err
	}
	if err := c.headerCell(sheet.Name, "A4", "C4", "D4", "H4", "Decimals:", fmt.Sprintf("%d", sheet.Decimals)); err != nil {
		return err
	}

	selections := []excelize.Selection{}

	c.setStyle(sheet.Name, fmt.Sprintf("A%d", headerRow), fmt.Sprintf((lastCol+"%d"), headerRow), styles.tableHeader)
	c.setStyle(sheet.Name, fmt.Sprintf("A%d", 1), fmt.Sprintf("F%d", 4), styles.mainHeader)
	c.setStyle(sheet.Name, "D1", "D1", styles.link)
	c.setLink(sheet.Name, "D1", "https://etherscan.io/address/"+sheet.Address, "Open in Explorer")
	c.ExcelFile.SetPanes(sheet.Name, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      headerRow,
		TopLeftCell: "A" + fmt.Sprintf("%d", headerRow+1),
		ActivePane:  "bottomLeft",
		Selection: append(selections, excelize.Selection{
			SQRef:      "A" + fmt.Sprintf("%d", headerRow+1),
			ActiveCell: "A" + fmt.Sprintf("%d", headerRow+1),
			Pane:       "bottomLeft",
		}),
	})

	return nil
}

func (c *Excel) setLink(sheetName, cell, url, tooltip string) {
	if err := c.ExcelFile.SetCellHyperLink(sheetName, cell, url, "External", excelize.HyperlinkOpts{
		Tooltip: &tooltip,
	}); err != nil {
		panic(err)
	}
}

func (c *Excel) headerCell(sheetName, c1, c2, c3, c4, t, v string) (err error) {
	if err = c.ExcelFile.MergeCell(sheetName, c1, c2); err == nil {
		if err = c.ExcelFile.MergeCell(sheetName, c3, c4); err == nil {
			if err = c.ExcelFile.SetCellValue(sheetName, c1, t); err == nil {
				return c.ExcelFile.SetCellValue(sheetName, c3, v)
			}
		}
	}
	return
}
