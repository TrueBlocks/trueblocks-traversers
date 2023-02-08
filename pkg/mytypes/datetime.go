package mytypes

import (
	"fmt"
	"strings"
	"time"

	"github.com/bykof/gostradamus"
)

type DateTime struct {
	time.Time
}

func NewDateTime() DateTime {
	return DateTime{time.Now()}
}

func (dt *DateTime) String() string {
	ret := dt.Time.Format("2006-01-02T15:04:05")
	ret = strings.Replace(ret, "T", " ", -1)
	ret = strings.Replace(ret, "+0000", "", -1)
	return ret + " UTC"
}

func (dt *DateTime) MarshalCSV() (string, error) {
	return dt.String(), nil
}

func (dt *DateTime) UnmarshalCSV(csv string) (err error) {
	csv = strings.Replace(csv, " UTC", "", -1)
	csv = strings.Replace(csv, " ", "T", -1)
	fmt := strings.Replace(time.RFC3339, "Z07:00", "", -1)
	dt.Time, err = time.Parse(fmt, csv)
	return err
}

// func fromGostro(in gostradamus.DateTime) DateTime {
// 	return DateTime{in.Time()}
// }

func (dt *DateTime) EndOfMonth() DateTime {
	g := gostradamus.DateTimeFromTime(dt.Time).CeilMonth()
	t, _ := time.Parse("2006-01-02T15:04:05", g.Format("2006-01-02T15:04:05"))
	return DateTime{t}
}

func (dt *DateTime) EndOfYear() DateTime {
	g := gostradamus.DateTimeFromTime(dt.Time).CeilYear()
	t, _ := time.Parse("2006-01-02T15:04:05", g.Format("2006-01-02T15:04:05"))
	return DateTime{t}
}

func IsValidPeriod(period string) bool {
	switch period {
	case "current":
		fallthrough
	case "annually":
		fallthrough
	case "quarterly":
		fallthrough
	case "monthly":
		fallthrough
	case "weekly":
		fallthrough
	case "daily":
		fallthrough
	case "hourly":
		fallthrough
	case "minutely":
		fallthrough
	case "secondly":
		fallthrough
	case "dayofweek":
		fallthrough
	case "blockly":
		return true
	default:
		return false
	}
}

func GetDateKey(period string, date DateTime) string {
	switch period {
	case "current":
		return ""
	case "annually":
		return fmt.Sprintf("%04d", date.Year())
	case "quarterly":
		quarter := (date.Month()-1)/3 + 1
		return fmt.Sprintf("%04d Q%d", date.Year(), quarter)
	case "monthly":
		return fmt.Sprintf("%04d-%02d", date.Year(), date.Month())
	case "weekly":
		y, w := date.ISOWeek()
		return fmt.Sprintf("%04d-%02d", y, w)
	case "daily":
		return fmt.Sprintf("%04d-%02d-%02d", date.Year(), date.Month(), date.Day())
	case "hourly":
		return fmt.Sprintf("%04d-%02d-%02d %02d", date.Year(), date.Month(), date.Day(), date.Hour())
	case "minutely":
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d", date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute())
	case "secondly":
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second())
	case "dayofweek":
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		return dayNames[date.Weekday()]
	default:
		return fmt.Sprintf("%04d-%02d-%02d", date.Year(), date.Month(), date.Day())
	}
	// return ""
}
