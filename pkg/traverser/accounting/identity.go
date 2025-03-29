package accounting

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type Identity struct {
	Opts     traverser.Options
	NotFirst bool
	Count    uint64
}

func (c *Identity) Traverse(val *types.Statement) {
	c.Count++
	if c.NotFirst {
		fmt.Println(",")
	} else {
		fmt.Println("[")
	}
	c.NotFirst = true
	fmt.Println(val)
}

func (c *Identity) GetKey(r *types.Statement) string {
	return ""
}

func (c *Identity) Result() string {
	return "]"
}

func (a *Identity) Name() string {
	return colors.Green + reflect.TypeOf(a).Elem().String() + colors.Off + ": " + fmt.Sprintf("%d", a.Count)
}

func (c *Identity) Sort(array []*types.Statement) {
	// Nothing to do
}
