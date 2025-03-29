package accounting

import (
	"os"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/traverser"
	"github.com/go-test/deep"
)

func TestProfitAndLossBasic(t *testing.T) {
	colors.ColorsOff()
	opts := traverser.Options{
		Period: "daily",
		Denom:  "wei",
	}
	c := &ProfitAndLoss{
		Opts:    opts,
		Ledgers: make(map[string]*types.Statement),
	}
	c.Traverse(want)
	if len(c.Ledgers) != 1 {
		t.Errorf("Ledger count differs: got %d, want 1", len(c.Ledgers))
	}
	wantKey := "0x0000000000000000000000000000000000000001-ETH-0x000000000000000000000000000000000000000a-2025-03-26"
	if diff := deep.Equal(c.Ledgers[wantKey], want); diff != nil {
		t.Errorf("Statement differs: %v", diff)
	}
	if got := c.Result(); got != "" {
		t.Errorf("Result differs: got %q, want empty string", got)
	}
	if got := c.LastDate; got != "2025-03-26" {
		t.Errorf("LastDate differs: got %q, want \"2025-03-26\"", got)
	}
	if got := c.LastKey; got != wantKey {
		t.Errorf("LastKey differs: got %q, want %q", got, wantKey)
	}
}

func TestProfitAndLossTwoStatements(t *testing.T) {
	colors.ColorsOff()
	opts := traverser.Options{
		Period: "daily",
		Denom:  "wei",
	}
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = old
	}()
	c := &ProfitAndLoss{
		Opts:    opts,
		Ledgers: make(map[string]*types.Statement),
		w:       tabwriter.NewWriter(os.Stdout, 0, 0, 1, ',', 0),
	}
	c.Traverse(want)
	c.Traverse(want2)
	c.w.Flush()
	if len(c.Ledgers) != 2 {
		t.Errorf("Ledger count differs: got %d, want 2", len(c.Ledgers))
	}
	wantKey1 := "0x0000000000000000000000000000000000000001-ETH-0x000000000000000000000000000000000000000a-2025-03-26"
	wantKey2 := "0x0000000000000000000000000000000000000001-ETH-0x000000000000000000000000000000000000000a-2025-03-27"
	if diff := deep.Equal(c.Ledgers[wantKey1], want); diff != nil {
		t.Errorf("First statement differs: %v", diff)
	}
	if diff := deep.Equal(c.Ledgers[wantKey2], want2); diff != nil {
		t.Errorf("Second statement differs: %v", diff)
	}
	if got := c.Result(); got != "" {
		t.Errorf("Result differs: got %q, want empty string", got)
	}
	if got := c.LastDate; got != "2025-03-27" {
		t.Errorf("LastDate differs: got %q, want \"2025-03-27\"", got)
	}
	if got := c.LastKey; got != wantKey2 {
		t.Errorf("LastKey differs: got %q, want %q", got, wantKey2)
	}
}

var want = &types.Statement{
	AccountedFor:        base.HexToAddress("0xa"),
	AmountIn:            *base.NewWei(1000),
	AmountOut:           *base.NewWei(0),
	Asset:               base.HexToAddress("0x1"),
	BegBal:              *base.NewWei(0),
	BlockNumber:         100,
	CorrectAmountIn:     *base.NewWei(0),
	CorrectAmountOut:    *base.NewWei(0),
	CorrectBegBalIn:     *base.NewWei(0),
	CorrectBegBalOut:    *base.NewWei(0),
	CorrectEndBalIn:     *base.NewWei(0),
	CorrectEndBalOut:    *base.NewWei(0),
	CorrectingReasons:   "",
	Decimals:            18,
	EndBal:              *base.NewWei(1000),
	GasOut:              *base.NewWei(0),
	InternalIn:          *base.NewWei(0),
	InternalOut:         *base.NewWei(0),
	LogIndex:            0,
	MinerBaseRewardIn:   *base.NewWei(0),
	MinerNephewRewardIn: *base.NewWei(0),
	MinerTxFeeIn:        *base.NewWei(0),
	MinerUncleRewardIn:  *base.NewWei(0),
	PrefundIn:           *base.NewWei(0),
	PrevBal:             *base.NewWei(0),
	PriceSource:         "",
	Recipient:           base.Address{},
	SelfDestructIn:      *base.NewWei(0),
	SelfDestructOut:     *base.NewWei(0),
	Sender:              base.Address{},
	SpotPrice:           *base.NewFloat(2000.0),
	Symbol:              "ETH",
	Timestamp:           timestamp,
	TransactionHash:     base.Hash{},
	TransactionIndex:    0,
	CorrectionId:        0,
	Holder:              base.Address{},
	StatementId:         0,
	BegSentinel:         false,
	EndSentinel:         false,
	Transaction:         nil,
	Log:                 nil,
}

var fixedTime = time.Date(2025, 3, 26, 0, 0, 0, 0, time.UTC)
var timestamp = base.Timestamp(fixedTime.Unix())

var want2 = &types.Statement{
	AccountedFor:        base.HexToAddress("0xa"),
	AmountIn:            *base.NewWei(0),
	AmountOut:           *base.NewWei(500),
	Asset:               base.HexToAddress("0x1"),
	BegBal:              *base.NewWei(1000),
	BlockNumber:         101,
	CorrectAmountIn:     *base.NewWei(0),
	CorrectAmountOut:    *base.NewWei(0),
	CorrectBegBalIn:     *base.NewWei(0),
	CorrectBegBalOut:    *base.NewWei(0),
	CorrectEndBalIn:     *base.NewWei(0),
	CorrectEndBalOut:    *base.NewWei(0),
	CorrectingReasons:   "",
	Decimals:            18,
	EndBal:              *base.NewWei(500),
	GasOut:              *base.NewWei(0),
	InternalIn:          *base.NewWei(0),
	InternalOut:         *base.NewWei(0),
	LogIndex:            0,
	MinerBaseRewardIn:   *base.NewWei(0),
	MinerNephewRewardIn: *base.NewWei(0),
	MinerTxFeeIn:        *base.NewWei(0),
	MinerUncleRewardIn:  *base.NewWei(0),
	PrefundIn:           *base.NewWei(0),
	PrevBal:             *base.NewWei(0),
	PriceSource:         "",
	Recipient:           base.Address{},
	SelfDestructIn:      *base.NewWei(0),
	SelfDestructOut:     *base.NewWei(0),
	Sender:              base.Address{},
	SpotPrice:           *base.NewFloat(2100.0),
	Symbol:              "ETH",
	Timestamp:           timestamp2,
	TransactionHash:     base.Hash{},
	TransactionIndex:    0,
	CorrectionId:        0,
	Holder:              base.Address{},
	StatementId:         0,
	BegSentinel:         false,
	EndSentinel:         false,
	Transaction:         nil,
	Log:                 nil,
}

var fixedTime2 = time.Date(2025, 3, 27, 0, 0, 0, 0, time.UTC)
var timestamp2 = base.Timestamp(fixedTime2.Unix())
