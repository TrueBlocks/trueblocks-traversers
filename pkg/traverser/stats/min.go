package stats

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type Min struct {
	Opts  traverser.Options
	Value float64
}

func (c *Min) Traverse(val float64) {
	if val < c.Value {
		c.Value = val
	}
}

func (c *Min) GetKey(unused float64) string {
	return ""
}

func (c *Min) Result() string {
	return c.Name() + "\n\t" + c.reportValue("Min: ", c.Value)
}

func (c *Min) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *Min) Sort(array []float64) {
	// Nothing to do
}

func (c *Min) reportValue(msg string, v float64) string {
	return fmt.Sprintf("%s%f", msg, v)
}
