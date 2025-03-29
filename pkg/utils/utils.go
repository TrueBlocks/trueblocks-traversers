package utils

import (
	"math/big"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
)

func PriceUsd(amtInt string, decimals base.Value, spotPrice *big.Float) *big.Float {
	// fmt.Println(amtInt, decimals, spotPrice)
	amount, _ := new(big.Float).SetString(amtInt)
	if amount == Zero() || decimals == 0 {
		z := Zero()
		return z
	}
	ten := new(big.Float)
	ten = ten.SetFloat64(10)
	divisor := Pow(ten, uint64(decimals))
	if divisor == nil || divisor == Zero() {
		return nil
	}
	units := Zero().Quo(amount, divisor)
	usd := Zero().Mul(units, spotPrice)
	usd.SetMode(big.ToNearestEven)
	usd.SetPrec(62) // gives us 19 digits of precision
	return usd
}

func Pow(a *big.Float, e uint64) *big.Float {
	result := Zero().Copy(a)
	for i := uint64(0); i < e-1; i++ {
		result = Zero().Mul(result, a)
	}
	return result
}

func Zero() *big.Float {
	r := big.NewFloat(0.0)
	r.SetPrec(256)
	return r
}
