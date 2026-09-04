package service

import (
	"errors"
	"math"
)

// Errors caused by the caller's input rather than by a fault in the service.
// The handler maps these to 400; anything else stays a 500.
var (
	ErrDivideByZero       = errors.New("cannot divide by zero")
	ErrNegativeSquareRoot = errors.New("cannot take the square root of a negative number")
)

type CalculatorService interface {
	Sum(first float64, second float64) (float64, error)
	Subtract(first float64, second float64) (float64, error)
	Division(first float64, second float64) (float64, error)
	Multiply(first float64, second float64) (float64, error)
	Exponentiation(base float64, exponent float64) (float64, error)
	SquareRoot(value float64) (float64, error)
	Percentage(value float64) (float64, error)
}

type CalculatorServiceImpl struct{}

func (cs *CalculatorServiceImpl) Sum(first float64, second float64) (float64, error) {
	return first + second, nil
}

func (cs *CalculatorServiceImpl) Subtract(first float64, second float64) (float64, error) {
	return first - second, nil
}

func (cs *CalculatorServiceImpl) Multiply(first float64, second float64) (float64, error) {
	return first * second, nil
}

func (cs *CalculatorServiceImpl) Division(first float64, second float64) (float64, error) {
	if second == 0 {
		return 0, ErrDivideByZero
	}
	return first / second, nil
}

func (cs *CalculatorServiceImpl) Exponentiation(base float64, exponent float64) (float64, error) {
	return math.Pow(base, exponent), nil
}

func (cs *CalculatorServiceImpl) SquareRoot(value float64) (float64, error) {
	if value < 0 {
		return 0, ErrNegativeSquareRoot
	}
	return math.Sqrt(value), nil
}

func (cs *CalculatorServiceImpl) Percentage(value float64) (float64, error) {
	return value / 100, nil
}

func NewCalculatorService() CalculatorService {
	return &CalculatorServiceImpl{}
}
