package accounting

import (
	"math/big"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-traversers/pkg/utils"
)

func ToUnits(amount *big.Float, decimals base.Value) *big.Float {
	if amount == utils.Zero() || decimals == 0 {
		return utils.Zero()
	}
	ten := new(big.Float)
	ten = ten.SetFloat64(10)
	divisor := utils.Pow(ten, uint64(decimals))
	if divisor == nil || divisor == utils.Zero() {
		return nil
	}
	return utils.Zero().Quo(amount, divisor)
}
