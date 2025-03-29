package accounting

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
)

func (c *Excel) SetCell(sheetName string, row int, sumRange CellRange, field *Field, val interface{}) string {
	var err error
	cell := fmt.Sprintf("%s%d", field.Column, row)
	switch field.Format {
	case "int":
		err = c.ExcelFile.SetCellInt(sheetName, cell, val.(int))
	case "formula":
		f := field.Formula
		f = strings.Replace(f, "{R}", strconv.Itoa(row), -1)
		if strings.Contains(f, "{R-1}") {
			if row == headerRow+1 {
				f = "=0"
			} else {
				f = strings.Replace(f, "{R-1}", strconv.Itoa(row-1), -1)
			}
		}
		f = strings.Replace(f, "{A}", strconv.Itoa(sumRange.A+1), -1)
		f = strings.Replace(f, "{B}", strconv.Itoa(sumRange.B-1), -1)
		if strings.Contains(f, "{L") && len(sumRange.CurRows) > 0 {
			if strings.Contains(f, "{L}") {
				sum := ""
				for i, v := range sumRange.CurRows {
					if i > 0 {
						sum += "+"
					}
					sum += fmt.Sprintf("%s%d", field.Column, v)
				}
				f = strings.Replace(f, "{L}", sum, -1)
			} else {
				f = strings.Replace(f, "{L0}", strconv.Itoa(sumRange.CurRows[0]), -1)
				f = strings.Replace(f, "{LN-1}", strconv.Itoa(sumRange.CurRows[len(sumRange.CurRows)-1]), -1)
			}
		}
		err = c.ExcelFile.SetCellFormula(sheetName, cell, f)
	case "float2":
		switch v := val.(type) {
		case float64:
			err = c.ExcelFile.SetCellFloat(sheetName, cell, v, 2, 64)
		case base.Float:
			err = c.ExcelFile.SetCellFloat(sheetName, cell, v.Float64(), 2, 64)
		}
	case "float5":
		switch v := val.(type) {
		case float64:
			err = c.ExcelFile.SetCellFloat(sheetName, cell, v, 5, 64)
		case base.Float:
			err = c.ExcelFile.SetCellFloat(sheetName, cell, v.Float64(), 5, 64)
		}
	case "bool":
		err = c.ExcelFile.SetCellBool(sheetName, cell, val.(bool))
	case "big":
		switch v := val.(type) {
		case string:
			var x big.Float
			x.SetString(v)
			f, _ := x.Float64()
			err = c.ExcelFile.SetCellFloat(sheetName, cell, f, 18, 64)
		case base.Wei:
			s := v.Text(10)
			var x big.Float
			x.SetString(s)
			f, _ := x.Float64()
			err = c.ExcelFile.SetCellFloat(sheetName, cell, f, 18, 64)
		}
	case "date":
		v := val.(base.DateTime)
		tt := time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond(), v.Location())
		err = c.ExcelFile.SetCellValue(sheetName, cell, tt)
	case "address":
		a := val.(base.Address)
		n := c.Opts.Names[base.HexToAddress(a.String())].Name
		if len(n) > 0 {
			if len(n) > 10 {
				n = n[0:10]
			}
			n = n + "-" + a.String()[0:min(6, len(a.String()))]
		} else {
			n = a.String()
		}
		err = c.ExcelFile.SetCellStr(sheetName, cell, n)
	case "hash":
		a := val.(base.Hash)
		s := a.String()
		err = c.ExcelFile.SetCellStr(sheetName, cell, s)
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
