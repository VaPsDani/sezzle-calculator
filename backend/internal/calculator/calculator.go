// Package calculator implements the arithmetic operations of the API.
package calculator

import (
	"errors"
	"math"
)

var (
	ErrDivisionByZero     = errors.New("division by zero")
	ErrNegativeSquareRoot = errors.New("square root of a negative number")
	ErrResultOutOfRange   = errors.New("result out of range")
)

func Add(a, b float64) float64 {
	return a + b
}

func Subtract(a, b float64) float64 {
	return a - b
}

func Multiply(a, b float64) float64 {
	return a * b
}

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

func Power(base, exponent float64) (float64, error) {
	return checkFinite(math.Pow(base, exponent))
}

func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSquareRoot
	}
	return checkFinite(math.Sqrt(a))
}

func Percentage(value, percent float64) float64 {
	return value * percent / 100
}

// Inf and NaN are valid float64 values but not valid answers, and NaN in
// particular propagates silently and compares false against everything.
func checkFinite(result float64) (float64, error) {
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, ErrResultOutOfRange
	}
	return result, nil
}
