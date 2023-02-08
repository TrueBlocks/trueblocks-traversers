package accounting

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/mytypes"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type Identity struct {
	Opts     traverser.Options
	NotFirst bool
	Count    uint64
}

func (c *Identity) Traverse(val *mytypes.RawReconciliation) {
	c.Count++
	if c.NotFirst {
		fmt.Println(",")
	} else {
		fmt.Println("[")
	}
	c.NotFirst = true
	fmt.Println(val)
}

func (c *Identity) GetKey(r *mytypes.RawReconciliation) string {
	return ""
}

func (c *Identity) Result() string {
	return "]"
}

func (a *Identity) Name() string {
	return colors.Green + reflect.TypeOf(a).Elem().String() + colors.Off + ": " + fmt.Sprintf("%d", a.Count)
}
