// Package calculator implements the arithmetic core of the application.
//
// It is deliberately free of any transport concern: it does not import
// net/http, encoding/json or any other web package. Callers translate its
// values and errors into whatever protocol they speak.
package calculator

import (
	"errors"
	"math"
)

// Sentinel errors returned by this package. Callers should test for them with
// errors.Is rather than comparing messages, so that the wording stays free to
// change and wrapped errors keep matching.
var (
	// ErrDivisionByZero is returned by Divide when the divisor is zero.
	ErrDivisionByZero = errors.New("division by zero")

	// ErrNegativeSquareRoot is returned by Sqrt when the operand is negative,
	// which has no real-valued square root.
	ErrNegativeSquareRoot = errors.New("square root of a negative number")

	// ErrResultOutOfRange is returned when an operation produces a value that
	// float64 cannot represent, such as an overflow to infinity or an
	// undefined result.
	ErrResultOutOfRange = errors.New("result out of range")
)

// Add returns the sum of a and b.
func Add(a, b float64) float64 {
	return a + b
}

// Subtract returns the difference of a and b.
func Subtract(a, b float64) float64 {
	return a - b
}

// Multiply returns the product of a and b.
func Multiply(a, b float64) float64 {
	return a * b
}

// Divide returns the quotient of a and b.
// It returns ErrDivisionByZero if b is zero, in which case the returned
// float64 is the zero value and carries no meaning.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Power returns base raised to exponent.
// It returns ErrResultOutOfRange when the result overflows the range of
// float64 or is undefined, for example 1e308 squared, or a negative base
// raised to a fractional exponent.
func Power(base, exponent float64) (float64, error) {
	return checkFinite(math.Pow(base, exponent))
}

// Sqrt returns the square root of a.
// It returns ErrNegativeSquareRoot if a is negative, since the result would
// not be a real number, and ErrResultOutOfRange if the result is not finite.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSquareRoot
	}
	return checkFinite(math.Sqrt(a))
}

// Percentage returns percent percent of value, so Percentage(200, 10) is 20.
func Percentage(value, percent float64) float64 {
	return value * percent / 100
}

// checkFinite guards the result of a computation against values that are
// representable in the IEEE-754 encoding but meaningless as an answer.
//
// It rejects both infinities, produced by overflow such as 1e308 squared, and
// NaN, produced by undefined operations such as math.Pow(-8, 1.0/3.0). On
// rejection it returns the zero value so that callers never receive a poisoned
// number alongside an error: NaN propagates silently through later arithmetic
// and compares false against everything, including itself.
func checkFinite(result float64) (float64, error) {
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, ErrResultOutOfRange
	}
	return result, nil
}
