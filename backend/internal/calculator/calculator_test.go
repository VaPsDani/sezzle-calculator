package calculator

import (
	"errors"
	"math"
	"testing"
)

// epsilon is the tolerance used to compare floating point results.
// Binary floating point cannot represent most decimal fractions exactly,
// so results are compared for closeness instead of equality.
//
// This is an absolute tolerance, which is appropriate because every value
// under test is small. Comparing magnitudes near the limits of float64
// would call for a relative tolerance instead.
const epsilon = 1e-9

// almostEqual reports whether got and want are within epsilon of each other.
func almostEqual(got, want float64) bool {
	return math.Abs(got-want) <= epsilon
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
