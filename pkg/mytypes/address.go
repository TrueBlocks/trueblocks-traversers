package mytypes

import (
	"strings"
)

type Address string

func (a *Address) String() string {
	if a.IsZero() {
		return "0x0000000000000000000000000000000000000000"
	}
	ret := strings.ToLower(string(*a))
	return ret
}

func (a *Address) UnmarshalCSV(csv string) (err error) {
	*a = *(*Address)(&csv)
	return nil
}

func (a *Address) IsZero() bool {
	return string(*a) == "0x0"
}
