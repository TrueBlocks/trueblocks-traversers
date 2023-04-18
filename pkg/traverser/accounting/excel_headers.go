package accounting

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuri/excelize/v2"
)

func (c *Excel) SetHeader(sheet *AssetSheet, styles *Styles, lastCol string) error {
	if err := c.headerCell(sheet.Name, "A1", "C1", "D1", "H1", "Asset Address:", sheet.Address); err != nil {
		return err
	}

	name := c.Opts.Names[common.HexToAddress(sheet.Address)].Name
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

	link, tooltip := "https://etherscan.io/address/"+sheet.Address, "Open in Explorer"
	if err := c.ExcelFile.SetCellHyperLink(sheet.Name, "D1", link, "External", excelize.HyperlinkOpts{
		Tooltip: &tooltip,
	}); err != nil {
		return err
	}

	c.SetStyle(sheet.Name, "D1", "D1", styles.link)
	c.SetStyle(sheet.Name, fmt.Sprintf("A%d", headerRow), fmt.Sprintf((lastCol+"%d"), headerRow), styles.tableHeader)
	c.SetStyle(sheet.Name, fmt.Sprintf("A%d", 1), fmt.Sprintf("F%d", 4), styles.mainHeader)

	return nil
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
