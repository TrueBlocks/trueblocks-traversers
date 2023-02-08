package utils

import (
	"math/big"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
)

func PriceUsd(amtInt string, decimals uint64, spotPrice *big.Float) *big.Float {
	// fmt.Println(amtInt, decimals, spotPrice)
	amount, _ := new(big.Float).SetString(amtInt)
	if amount == Zero() || decimals == 0 {
		z := Zero()
		return z
	}
	ten := new(big.Float)
	ten = ten.SetFloat64(10)
	divisor := Pow(ten, decimals)
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

func ColorsOn() {
	colors.Off = "\033[0m"
	colors.Red = "\033[31m"
	colors.Green = "\033[32m"
	colors.Yellow = "\033[33m"
	colors.Blue = "\033[34m"
	colors.Magenta = "\033[35m"
	colors.Cyan = "\033[36m"
	colors.White = "\033[37m"
	colors.Black = "\033[30m"
	colors.Bright = "\033[1m"
	colors.BrightRed = (colors.Bright + colors.Red)
	colors.BrightGreen = (colors.Bright + colors.Green)
	colors.BrightYellow = (colors.Bright + colors.Yellow)
	colors.BrightBlue = (colors.Bright + colors.Blue)
	colors.BrightMagenta = (colors.Bright + colors.Magenta)
	colors.BrightCyan = (colors.Bright + colors.Cyan)
	colors.BrightWhite = (colors.Bright + colors.White)
	colors.BrightBlack = (colors.Bright + colors.Black)
}
