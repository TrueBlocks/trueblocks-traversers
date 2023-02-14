package excel

import (
	"github.com/xuri/excelize/v2"
)

func NewWorkbook(sheetName string, header []string) *excelize.File {
	f := excelize.NewFile()
	f.WorkBook.BookViews.WorkBookView[0].WindowHeight = 0
	f.WorkBook.BookViews.WorkBookView[0].WindowWidth = 0
	f.SetSheetName("Sheet1", sheetName)
	if len(header) > 0 {
		f.SetCellValue(sheetName, "A1", header[0])
	}

	return f
}

type Style int
type Styles []Style

func GetStyles() []Styles {
	return []Styles{
		{0},
	}
}
