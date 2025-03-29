package logs

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type Counter struct {
	Opts  traverser.Options
	Value uint64
}

func (c *Counter) Traverse(r *types.Log) {
	c.Value += 1
}

func (c *Counter) GetKey(r *types.Log) string {
	return ""
}

func (c *Counter) Result() string {
	return c.Name() + "\n\t" + c.reportValue("Counter: ", c.Value)
}

func (c *Counter) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *Counter) Sort(array []*types.Log) {
	// Nothing to do
}

func (c *Counter) reportValue(msg string, v uint64) string {
	return fmt.Sprintf("%s%d", msg, v)
}
