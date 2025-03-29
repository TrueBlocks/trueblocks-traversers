package logs

import (
	"fmt"
	"reflect"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
)

// --------------------------------
type ExtractLog struct {
	Opts  traverser.Options
	Count uint64
	// w     *tabwriter.Writer
}

func (c *ExtractLog) Traverse(l *types.Log) {
	if c.Count == 0 {
		// c.w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ',', 0)
		c.ReportHeader(c.Opts.Verbose, l)
	}
	c.ReportRecord(l)
	c.Count++
}

func (c *ExtractLog) GetKey(r *types.Log) string {
	return ""
}

func (c *ExtractLog) Result() string {
	return ""
}

func (a *ExtractLog) Name() string {
	return colors.Green + reflect.TypeOf(a).Elem().String() + colors.Off + ": " + fmt.Sprintf("%d", a.Count)
}

func (c *ExtractLog) Sort(array []*types.Log) {
	// Nothing to do
}

func (c *ExtractLog) ReportHeader(verbose int, r *types.Log) {
	if verbose > 0 {
		fmt.Println("Block\tTx\tLog\tAddress\tTopics\tData\t")
	}
}

func (c *ExtractLog) ReportRecord(r *types.Log) {
	// if c.Count%20 == 0 {
	// 	c.w.Flush()
	// }
	fmt.Printf("%d\t%d\t%d\t%s\t%s\n", r.BlockNumber, r.TransactionIndex, r.LogIndex, r.Address.Hex(), r.CompressedLog())
}
