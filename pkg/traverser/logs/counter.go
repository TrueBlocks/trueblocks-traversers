package logs

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type Counter struct {
	Opts  traverser.Options
	Value uint64
}

func (c *Counter) Traverse(r *mytypes.RawLog) {
	c.Value += 1
}

func (c *Counter) GetKey(r *mytypes.RawLog) string {
	return ""
}

func (c *Counter) Result() string {
	return c.Name() + "\n\t" + c.reportValue("Counter: ", c.Value)
}

func (c *Counter) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *Counter) reportValue(msg string, v uint64) string {
	return fmt.Sprintf("%s%d", msg, v)
}
