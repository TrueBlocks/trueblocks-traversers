package stats

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

func TestCounter(t *testing.T) {
	tests := []struct {
		name       string
		inputs     []float64
		wantCount  float64
		wantResult string
	}{
		{"Empty", []float64{}, 0, "stats.Counter\n\tCounter: 0.000000"},
		{"Single", []float64{1.0}, 1, "stats.Counter\n\tCounter: 1.000000"},
		{"Multiple", []float64{1.0, 2.0, 3.0}, 3, "stats.Counter\n\tCounter: 3.000000"},
	}

	colors.ColorsOff()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Counter{Opts: traverser.Options{}}
			for _, val := range tt.inputs {
				c.Traverse(val)
			}
			if c.Value != tt.wantCount {
				t.Errorf("Traverse failed: got %f, want %f", c.Value, tt.wantCount)
			}
			if got := c.Result(); got != tt.wantResult {
				t.Errorf("Result failed: got %q, want %q", got, tt.wantResult)
			}
		})
	}
}
