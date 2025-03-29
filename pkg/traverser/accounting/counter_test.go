package accounting

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

func TestCounter(t *testing.T) {
	colors.ColorsOff()
	c := &Counter{Opts: traverser.Options{}}
	stmt1 := &types.Statement{BlockNumber: 1}
	stmt2 := &types.Statement{BlockNumber: 2}

	c.Traverse(stmt1)
	c.Traverse(stmt2)

	if c.Value != 2 {
		t.Errorf("Traverse failed: got %d, want 2", c.Value)
	}
	wantResult := "accounting.Counter\n\tCounter: 2"
	if got := c.Result(); got != wantResult {
		t.Errorf("Result failed: got %q, want %q", got, wantResult)
	}
}
