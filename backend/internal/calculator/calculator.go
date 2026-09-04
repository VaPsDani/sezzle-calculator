// Package calculator implements the arithmetic core of the application.
//
// It is deliberately free of any transport concern: it does not import
// net/http, encoding/json or any other web package. Callers translate its
// values and errors into whatever protocol they speak.
package calculator

import "errors"

// ErrDivisionByZero is returned by Divide when the divisor is zero.
// Callers should test for it with errors.Is rather than comparing messages.
var ErrDivisionByZero = errors.New("division by zero")

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
