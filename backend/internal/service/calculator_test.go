package service

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type binaryCase struct {
	name    string
	a, b    float64
	want    float64
	wantErr error
}

func runBinaryCases(t *testing.T, op func(a, b float64) (float64, error), cases []binaryCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := op(tc.a, tc.b)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Zero(t, got)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tc.want, got, 1e-9)
			assert.False(t, math.Signbit(got) && got == 0, "result must not be negative zero")
		})
	}
}

// invalidOperandCases are shared by every binary operation.
func invalidOperandCases() []binaryCase {
	return []binaryCase{
		{name: "NaN operand", a: math.NaN(), b: 1, wantErr: ErrInvalidOperand},
		{name: "+Inf operand", a: 1, b: math.Inf(1), wantErr: ErrInvalidOperand},
		{name: "-Inf operand", a: math.Inf(-1), b: 1, wantErr: ErrInvalidOperand},
	}
}

func TestAdd(t *testing.T) {
	svc := NewCalculatorService()
	runBinaryCases(t, svc.Add, append([]binaryCase{
		{name: "positive integers", a: 10, b: 4, want: 14},
		{name: "negative and positive", a: -10, b: 4, want: -6},
		{name: "decimals", a: 0.1, b: 0.2, want: 0.3},
		{name: "zeros", a: 0, b: 0, want: 0},
		{name: "negative zero normalized", a: math.Copysign(0, -1), b: math.Copysign(0, -1), want: 0},
		{name: "overflow", a: math.MaxFloat64, b: math.MaxFloat64, wantErr: ErrResultOutOfRange},
	}, invalidOperandCases()...))
}

func TestSubtract(t *testing.T) {
	svc := NewCalculatorService()
	runBinaryCases(t, svc.Subtract, append([]binaryCase{
		{name: "positive result", a: 10, b: 4, want: 6},
		{name: "negative result", a: 4, b: 10, want: -6},
		{name: "same numbers", a: 2.5, b: 2.5, want: 0},
		{name: "overflow", a: -math.MaxFloat64, b: math.MaxFloat64, wantErr: ErrResultOutOfRange},
	}, invalidOperandCases()...))
}

func TestMultiply(t *testing.T) {
	svc := NewCalculatorService()
	runBinaryCases(t, svc.Multiply, append([]binaryCase{
		{name: "positive integers", a: 10, b: 4, want: 40},
		{name: "negative times positive", a: -3, b: 4, want: -12},
		{name: "by zero", a: 123.45, b: 0, want: 0},
		{name: "negative times zero is plain zero", a: -5, b: 0, want: 0},
		{name: "decimals", a: 1.5, b: 1.5, want: 2.25},
		{name: "overflow", a: 1e200, b: 1e200, wantErr: ErrResultOutOfRange},
	}, invalidOperandCases()...))
}

func TestDivide(t *testing.T) {
	svc := NewCalculatorService()
	runBinaryCases(t, svc.Divide, append([]binaryCase{
		{name: "exact", a: 10, b: 4, want: 2.5},
		{name: "repeating decimal", a: 1, b: 3, want: 1.0 / 3},
		{name: "negative divisor", a: 10, b: -4, want: -2.5},
		{name: "zero dividend", a: 0, b: 5, want: 0},
		{name: "division by zero", a: 10, b: 0, wantErr: ErrDivisionByZero},
		{name: "zero by zero", a: 0, b: 0, wantErr: ErrDivisionByZero},
		{name: "division by negative zero", a: 10, b: math.Copysign(0, -1), wantErr: ErrDivisionByZero},
		{name: "overflow", a: math.MaxFloat64, b: 0.5, wantErr: ErrResultOutOfRange},
	}, invalidOperandCases()...))
}

func TestPower(t *testing.T) {
	svc := NewCalculatorService()
	runBinaryCases(t, svc.Power, append([]binaryCase{
		{name: "integer exponent", a: 2, b: 10, want: 1024},
		{name: "zero exponent", a: 5, b: 0, want: 1},
		{name: "negative exponent", a: 2, b: -2, want: 0.25},
		{name: "fractional exponent", a: 9, b: 0.5, want: 3},
		{name: "negative base, integer exponent", a: -2, b: 3, want: -8},
		{name: "overflow", a: 10, b: 400, wantErr: ErrResultOutOfRange},
		{name: "zero to a negative power", a: 0, b: -1, wantErr: ErrResultOutOfRange},
		{name: "negative base, fractional exponent", a: -8, b: 0.5, wantErr: ErrUndefinedResult},
	}, invalidOperandCases()...))
}

func TestPercentage(t *testing.T) {
	svc := NewCalculatorService()
	runBinaryCases(t, svc.Percentage, append([]binaryCase{
		{name: "15 percent of 200", a: 15, b: 200, want: 30},
		{name: "100 percent", a: 100, b: 42, want: 42},
		{name: "decimal percent", a: 12.5, b: 80, want: 10},
		{name: "negative percent", a: -10, b: 50, want: -5},
		{name: "zero percent", a: 0, b: 50, want: 0},
		{name: "large operands avoid intermediate overflow", a: 50, b: math.MaxFloat64, want: math.MaxFloat64 / 2},
		{name: "overflow", a: 1e300, b: 1e300, wantErr: ErrResultOutOfRange},
	}, invalidOperandCases()...))
}

func TestPercentageKeepsIntegerInputsExact(t *testing.T) {
	got, err := NewCalculatorService().Percentage(15, 200)
	require.NoError(t, err)
	assert.Equal(t, 30.0, got) // exact, not just within a delta
}

func TestSqrt(t *testing.T) {
	svc := NewCalculatorService()
	cases := []struct {
		name    string
		a       float64
		want    float64
		wantErr error
	}{
		{name: "perfect square", a: 16, want: 4},
		{name: "non-perfect square", a: 2, want: math.Sqrt2},
		{name: "zero", a: 0, want: 0},
		{name: "negative zero", a: math.Copysign(0, -1), want: 0},
		{name: "fraction", a: 0.25, want: 0.5},
		{name: "negative", a: -4, wantErr: ErrNegativeSquareRoot},
		{name: "NaN", a: math.NaN(), wantErr: ErrInvalidOperand},
		{name: "+Inf", a: math.Inf(1), wantErr: ErrInvalidOperand},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.Sqrt(tc.a)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tc.want, got, 1e-9)
			assert.False(t, math.Signbit(got))
		})
	}
}
