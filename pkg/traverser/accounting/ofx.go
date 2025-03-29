package accounting

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type OfxWriter struct {
	Opts  traverser.Options
	Value uint64
}

func (c *OfxWriter) Traverse(r *types.Statement) {
	// SendToOfx(connection, r)
}

func (c *OfxWriter) GetKey(r *types.Statement) string {
	return ""
}

func (c *OfxWriter) Result() string {
	return c.Name() + "\n\t" + c.reportValue("OfxWriter: ", c.Value)
}

func (c *OfxWriter) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *OfxWriter) reportValue(msg string, v uint64) string {
	return fmt.Sprintf("%s%d", msg, v)
}
