package calculator

import (
	"errors"
	"math"
	"testing"
)

// epsilon is the tolerance used to compare floating point results.
// Binary floating point cannot represent most decimal fractions exactly,
// so results are compared for closeness instead of equality.
const epsilon = 1e-9

// almostEqual reports whether got and want are close enough to be treated as
// the same result.
//
// It combines an absolute and a relative tolerance. The absolute check covers
// values near zero, where a relative tolerance collapses to nothing. The
// relative check covers large magnitudes such as 1e154, where consecutive
// float64 values are already farther apart than epsilon and an absolute
// tolerance would be exactly as strict as ==.
func almostEqual(got, want float64) bool {
	diff := math.Abs(got - want)
	if diff <= epsilon {
		return true
	}
	return diff <= epsilon*math.Max(math.Abs(got), math.Abs(want))
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{name: "positive integers", a: 2, b: 3, want: 5},
		{name: "negative operands", a: -7, b: -3, want: -10},
		{name: "mixed signs", a: -8, b: 3, want: -5},
		{name: "decimals", a: 0.1, b: 0.2, want: 0.3},
		{name: "decimals with negatives", a: -2.75, b: 1.25, want: -1.5},
		{name: "zero as left operand", a: 0, b: 42, want: 42},
		{name: "zero as right operand", a: 42, b: 0, want: 42},
		{name: "both operands zero", a: 0, b: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if !almostEqual(got, tt.want) {
				t.Errorf("Add(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{name: "positive integers", a: 9, b: 4, want: 5},
		{name: "result goes negative", a: 4, b: 9, want: -5},
		{name: "negative operands", a: -7, b: -3, want: -4},
		{name: "mixed signs", a: -8, b: 3, want: -11},
		{name: "decimals", a: 0.3, b: 0.1, want: 0.2},
		{name: "decimals with negatives", a: -2.75, b: 1.25, want: -4},
		{name: "zero as left operand", a: 0, b: 42, want: -42},
		{name: "zero as right operand", a: 42, b: 0, want: 42},
		{name: "both operands zero", a: 0, b: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			if !almostEqual(got, tt.want) {
				t.Errorf("Subtract(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{name: "positive integers", a: 6, b: 7, want: 42},
		{name: "negative times positive", a: -6, b: 7, want: -42},
		{name: "negative times negative", a: -6, b: -7, want: 42},
		{name: "decimals", a: 1.1, b: 3, want: 3.3},
		{name: "decimals with negatives", a: -2.5, b: 0.4, want: -1},
		{name: "zero as left operand", a: 0, b: 42, want: 0},
		{name: "zero as right operand", a: 42, b: 0, want: 0},
		{name: "both operands zero", a: 0, b: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Multiply(tt.a, tt.b)
			if !almostEqual(got, tt.want) {
				t.Errorf("Multiply(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr error // nil means the call is expected to succeed
	}{
		{name: "positive integers", a: 10, b: 2, want: 5},
		{name: "non exact quotient", a: 10, b: 4, want: 2.5},
		{name: "negative divided by positive", a: -10, b: 2, want: -5},
		{name: "negative divided by negative", a: -10, b: -2, want: 5},
		{name: "decimals", a: 7.5, b: 2.5, want: 3},
		{name: "decimals with negatives", a: -0.75, b: 0.25, want: -3},
		{name: "zero as dividend", a: 0, b: 42, want: 0},
		{name: "zero as divisor", a: 42, b: 0, wantErr: ErrDivisionByZero},
		{name: "negative zero as divisor", a: 42, b: math.Copysign(0, -1), wantErr: ErrDivisionByZero},
		{name: "zero divided by zero", a: 0, b: 0, wantErr: ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Divide(%v, %v) error = %v, want %v", tt.a, tt.b, err, tt.wantErr)
				}
				// On failure the numeric result carries no meaning and must be the zero value.
				if got != 0 {
					t.Errorf("Divide(%v, %v) = %v on error, want 0", tt.a, tt.b, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if !almostEqual(got, tt.want) {
				t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestPower(t *testing.T) {
	tests := []struct {
		name           string
		base, exponent float64
		want           float64
		wantErr        error // nil means the call is expected to succeed
	}{
		{name: "positive integers", base: 2, exponent: 3, want: 8},
		{name: "negative base with even exponent", base: -2, exponent: 2, want: 4},
		{name: "negative base with odd exponent", base: -2, exponent: 3, want: -8},
		{name: "negative exponent", base: 2, exponent: -2, want: 0.25},
		{name: "decimal base", base: 1.5, exponent: 2, want: 2.25},
		{name: "decimal exponent is a root", base: 9, exponent: 0.5, want: 3},
		{name: "decimal base and exponent", base: 6.25, exponent: -0.5, want: 0.4},
		{name: "zero exponent", base: 5, exponent: 0, want: 1},
		{name: "zero base", base: 0, exponent: 5, want: 0},
		{name: "both operands zero", base: 0, exponent: 0, want: 1},
		{name: "underflow rounds to zero", base: 1e-200, exponent: 2, want: 0},
		{name: "largest finite result", base: 1e154, exponent: 2, want: 1e308},
		{name: "overflow to positive infinity", base: 1e308, exponent: 2, wantErr: ErrResultOutOfRange},
		{name: "overflow to negative infinity", base: -1e308, exponent: 3, wantErr: ErrResultOutOfRange},
		{name: "zero base with negative exponent", base: 0, exponent: -1, wantErr: ErrResultOutOfRange},
		{name: "undefined result is not a number", base: -8, exponent: 1.0 / 3.0, wantErr: ErrResultOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Power(tt.base, tt.exponent)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Power(%v, %v) error = %v, want %v", tt.base, tt.exponent, err, tt.wantErr)
				}
				// A rejected result must not leak Inf or NaN to the caller.
				if got != 0 {
					t.Errorf("Power(%v, %v) = %v on error, want 0", tt.base, tt.exponent, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Power(%v, %v) unexpected error: %v", tt.base, tt.exponent, err)
			}
			if !almostEqual(got, tt.want) {
				t.Errorf("Power(%v, %v) = %v, want %v", tt.base, tt.exponent, got, tt.want)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr error // nil means the call is expected to succeed
	}{
		{name: "perfect square", a: 9, want: 3},
		{name: "irrational result", a: 2, want: 1.4142135623730951},
		{name: "identity", a: 1, want: 1},
		{name: "decimals", a: 6.25, want: 2.5},
		{name: "decimal below one", a: 0.25, want: 0.5},
		{name: "zero", a: 0, want: 0},
		{name: "negative zero", a: math.Copysign(0, -1), want: 0},
		{name: "very large operand", a: 1e308, want: 1e154},
		{name: "negative integer", a: -9, wantErr: ErrNegativeSquareRoot},
		{name: "negative decimal", a: -0.25, wantErr: ErrNegativeSquareRoot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.a)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Sqrt(%v) error = %v, want %v", tt.a, err, tt.wantErr)
				}
				if got != 0 {
					t.Errorf("Sqrt(%v) = %v on error, want 0", tt.a, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Sqrt(%v) unexpected error: %v", tt.a, err)
			}
			if !almostEqual(got, tt.want) {
				t.Errorf("Sqrt(%v) = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		name           string
		value, percent float64
		want           float64
	}{
		{name: "positive integers", value: 200, percent: 10, want: 20},
		{name: "whole value", value: 50, percent: 100, want: 50},
		{name: "more than one hundred percent", value: 200, percent: 150, want: 300},
		{name: "negative percent", value: 200, percent: -10, want: -20},
		{name: "negative value", value: -200, percent: 10, want: -20},
		{name: "both operands negative", value: -200, percent: -10, want: 20},
		{name: "decimal percent", value: 80, percent: 12.5, want: 10},
		{name: "decimal value", value: 33.33, percent: 7.5, want: 2.49975},
		{name: "zero percent", value: 200, percent: 0, want: 0},
		{name: "zero value", value: 0, percent: 10, want: 0},
		{name: "both operands zero", value: 0, percent: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Percentage(tt.value, tt.percent)
			if !almostEqual(got, tt.want) {
				t.Errorf("Percentage(%v, %v) = %v, want %v", tt.value, tt.percent, got, tt.want)
			}
		})
	}
}
