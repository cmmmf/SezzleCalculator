// Package service holds the calculator business logic. It has no knowledge of
// HTTP, Gin, or logging, so it can be reused and tested in isolation.
package service

import (
	"errors"
	"math"
)

// Domain errors returned by CalculatorService. Callers should compare them
// with errors.Is.
var (
	// ErrInvalidOperand is returned when an operand is NaN or ±Inf.
	ErrInvalidOperand = errors.New("operands must be finite numbers")
	// ErrDivisionByZero is returned when dividing by zero.
	ErrDivisionByZero = errors.New("division by zero is not allowed")
	// ErrNegativeSquareRoot is returned for the square root of a negative number.
	ErrNegativeSquareRoot = errors.New("square root of a negative number is not a real number")
	// ErrResultOutOfRange is returned when the result overflows float64 (±Inf).
	ErrResultOutOfRange = errors.New("result is too large to be represented")
	// ErrUndefinedResult is returned when the result is not a real number (NaN),
	// e.g. a negative base raised to a fractional exponent.
	ErrUndefinedResult = errors.New("result is undefined for the given operands")
)

// CalculatorService performs arithmetic operations on float64 operands.
// Every method returns a finite result or one of the domain errors above.
type CalculatorService interface {
	Add(a, b float64) (float64, error)
	Subtract(a, b float64) (float64, error)
	Multiply(a, b float64) (float64, error)
	Divide(a, b float64) (float64, error)
	// Power returns a raised to the power of b.
	Power(a, b float64) (float64, error)
	// Percentage returns a percent of b (e.g. 15% of 200 = 30).
	Percentage(a, b float64) (float64, error)
	// Sqrt returns the square root of a.
	Sqrt(a float64) (float64, error)
}

type calculatorService struct{}

// NewCalculatorService returns the default CalculatorService implementation.
func NewCalculatorService() CalculatorService {
	return calculatorService{}
}

func (calculatorService) Add(a, b float64) (float64, error) {
	if err := validateOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(a + b)
}

func (calculatorService) Subtract(a, b float64) (float64, error) {
	if err := validateOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(a - b)
}

func (calculatorService) Multiply(a, b float64) (float64, error) {
	if err := validateOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(a * b)
}

func (calculatorService) Divide(a, b float64) (float64, error) {
	if err := validateOperands(a, b); err != nil {
		return 0, err
	}
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return checkResult(a / b)
}

func (calculatorService) Power(a, b float64) (float64, error) {
	if err := validateOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(math.Pow(a, b))
}

func (calculatorService) Percentage(a, b float64) (float64, error) {
	if err := validateOperands(a, b); err != nil {
		return 0, err
	}
	// Multiplying first keeps integer inputs exact (15*200/100 = 30, while
	// 15/100*200 = 30.000000000000004). Fall back to dividing first only when
	// the product itself overflows.
	result := a * b / 100
	if math.IsInf(result, 0) {
		result = a / 100 * b
	}
	return checkResult(result)
}

func (calculatorService) Sqrt(a float64) (float64, error) {
	if err := validateOperands(a); err != nil {
		return 0, err
	}
	if a < 0 {
		return 0, ErrNegativeSquareRoot
	}
	return checkResult(math.Sqrt(a))
}

func validateOperands(operands ...float64) error {
	for _, v := range operands {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return ErrInvalidOperand
		}
	}
	return nil
}

// checkResult rejects non-finite results and normalizes -0 to 0, so clients
// never see "-0" in a JSON response.
func checkResult(r float64) (float64, error) {
	switch {
	case math.IsNaN(r):
		return 0, ErrUndefinedResult
	case math.IsInf(r, 0):
		return 0, ErrResultOutOfRange
	case r == 0:
		return 0, nil
	default:
		return r, nil
	}
}
