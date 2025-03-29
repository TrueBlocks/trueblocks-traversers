package accounting

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type SqlWriter struct {
	Opts  traverser.Options
	Value uint64
}

func (c *SqlWriter) Traverse(r *types.Statement) {
	// SendToSQL(connection, r)
}

func (c *SqlWriter) GetKey(r *types.Statement) string {
	return ""
}

func (c *SqlWriter) Result() string {
	return c.Name() + "\n\t" + c.reportValue("SqlWriter: ", c.Value)
}

func (c *SqlWriter) Name() string {
	return colors.Green + reflect.TypeOf(c).Elem().String() + colors.Off
}

func (c *SqlWriter) reportValue(msg string, v uint64) string {
	return fmt.Sprintf("%s%d", msg, v)
}
