package excel

import (
	"github.com/xuri/excelize/v2"
)

func NewWorkbook(sheetName string, header []string) *excelize.File {
	f := excelize.NewFile()

	f.SetSheetName("Sheet1", sheetName)
	if len(header) > 0 {
		f.SetCellValue(sheetName, "A1", header[0])
	}

	return f
}
