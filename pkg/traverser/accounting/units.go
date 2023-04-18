package accounting

import (
	"math/big"

	"github.com/TrueBlocks/trueblocks-traversers/pkg/utils"
)

func ToUnits(amount *big.Float, decimals uint64) *big.Float {
	if amount == utils.Zero() || decimals == 0 {
		return utils.Zero()
	}
	ten := new(big.Float)
	ten = ten.SetFloat64(10)
	divisor := utils.Pow(ten, decimals)
	if divisor == nil || divisor == utils.Zero() {
		return nil
	}
	return utils.Zero().Quo(amount, divisor)
}
