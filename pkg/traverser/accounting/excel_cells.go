package accounting

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/ethereum/go-ethereum/common"
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
		f = strings.Replace(f, "{R-1}", strconv.Itoa(row-1), -1)
		f = strings.Replace(f, "{A}", strconv.Itoa(sumRange.A+1), -1)
		f = strings.Replace(f, "{B}", strconv.Itoa(sumRange.B-1), -1)
		if strings.Contains(f, "{L") && len(sumRange.CurRows) > 0 {
			f = strings.Replace(f, "{L0}", strconv.Itoa(sumRange.CurRows[0]), -1)
			f = strings.Replace(f, "{LN-1}", strconv.Itoa(sumRange.CurRows[len(sumRange.CurRows)-1]), -1)
			f = strings.Replace(f, "{L}", strconv.Itoa(1), -1)
		}
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
