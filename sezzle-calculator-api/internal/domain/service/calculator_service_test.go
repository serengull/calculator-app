package service

import (
	"errors"
	"math"
	"testing"
)

func TestCalculatorServiceImpl_Sum(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		first  float64
		second float64
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "should add two positive numbers",
			args: args{first: 2, second: 4},
			want: 6,
		},
		{
			name: "should add two negative numbers",
			args: args{first: -2, second: -4},
			want: -6,
		},
		{
			name: "should add a negative to a positive",
			args: args{first: 3, second: -7},
			want: -4,
		},
		{
			name: "should add decimals that binary floating point represents exactly",
			args: args{first: 1.5, second: 2.25},
			want: 3.75,
		},
		{
			// Documents the inherited float64 behaviour: the service does not
			// round, so the API returns 0.30000000000000004 for this input.
			name: "should carry binary floating point error",
			args: args{first: 0.1, second: 0.2},
			want: 0.30000000000000004,
		},
		{
			name: "should keep full precision at the edge of float64 integers",
			args: args{first: 9007199254740992, second: 1},
			want: 9007199254740992,
		},
		{
			name: "should lose an addend too small to change the result",
			args: args{first: 1, second: math.SmallestNonzeroFloat64},
			want: 1,
		},
		{
			name: "should overflow to positive infinity",
			args: args{first: math.MaxFloat64, second: math.MaxFloat64},
			want: math.Inf(1),
		},
		{
			name: "should overflow to negative infinity",
			args: args{first: -math.MaxFloat64, second: -math.MaxFloat64},
			want: math.Inf(-1),
		},
		{
			name: "should propagate an infinite operand",
			args: args{first: math.Inf(1), second: 1},
			want: math.Inf(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Sum(tt.args.first, tt.args.second)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if actual != tt.want {
				t.Errorf("Sum() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestCalculatorServiceImpl_Sum_ShouldPropagateNaN(t *testing.T) {
	service := NewCalculatorService()
	tests := []struct {
		name   string
		first  float64
		second float64
	}{
		{name: "should propagate a NaN operand", first: math.NaN(), second: 1},
		{name: "should propagate NaN on both sides", first: 1, second: math.NaN()},
		{name: "should return NaN for opposing infinities", first: math.Inf(1), second: math.Inf(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Sum(tt.first, tt.second)
			if err != nil {
				t.Errorf("Sum() unexpected error = %v", err)
				return
			}
			if !math.IsNaN(actual) {
				t.Errorf("Sum() got = %v, want NaN", actual)
			}
		})
	}
}

func TestCalculatorServiceImpl_Subtract(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		first  float64
		second float64
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "should subtract the second operand from the first",
			args: args{first: 9, second: 4},
			want: 5,
		},
		{
			// Argument order matters: the reverse would give 5.
			name: "should respect the argument order",
			args: args{first: 4, second: 9},
			want: -5,
		},
		{
			name: "should return the first operand when subtracting zero",
			args: args{first: 7, second: 0},
			want: 7,
		},
		{
			name: "should negate the second operand when subtracting from zero",
			args: args{first: 0, second: 7},
			want: -7,
		},
		{
			name: "should add when the second operand is negative",
			args: args{first: 5, second: -3},
			want: 8,
		},
		{
			name: "should subtract two negatives",
			args: args{first: -5, second: -3},
			want: -2,
		},
		{
			name: "should return zero for equal operands",
			args: args{first: 42.5, second: 42.5},
			want: 0,
		},
		{
			name: "should carry binary floating point error",
			args: args{first: 0.3, second: 0.1},
			want: 0.19999999999999998,
		},
		{
			name: "should overflow to negative infinity",
			args: args{first: -math.MaxFloat64, second: math.MaxFloat64},
			want: math.Inf(-1),
		},
		{
			name: "should overflow to positive infinity",
			args: args{first: math.MaxFloat64, second: -math.MaxFloat64},
			want: math.Inf(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Subtract(tt.args.first, tt.args.second)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subtract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if actual != tt.want {
				t.Errorf("Subtract() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestCalculatorServiceImpl_Multiply(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		first  float64
		second float64
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "should multiply two positive numbers",
			args: args{first: 6, second: 7},
			want: 42,
		},
		{
			name: "should be commutative",
			args: args{first: 7, second: 6},
			want: 42,
		},
		{
			name: "should return zero when either operand is zero",
			args: args{first: 99, second: 0},
			want: 0,
		},
		{
			name: "should return the other operand when multiplying by one",
			args: args{first: 1, second: 42},
			want: 42,
		},
		{
			name: "should produce a negative for mixed signs",
			args: args{first: -6, second: 7},
			want: -42,
		},
		{
			name: "should produce a positive for two negatives",
			args: args{first: -6, second: -7},
			want: 42,
		},
		{
			name: "should multiply decimals",
			args: args{first: 1.5, second: 2.5},
			want: 3.75,
		},
		{
			name: "should carry binary floating point error",
			args: args{first: 0.1, second: 3},
			want: 0.30000000000000004,
		},
		{
			name: "should overflow to positive infinity",
			args: args{first: math.MaxFloat64, second: 2},
			want: math.Inf(1),
		},
		{
			name: "should overflow to negative infinity for mixed signs",
			args: args{first: math.MaxFloat64, second: -2},
			want: math.Inf(-1),
		},
		{
			name: "should underflow to zero",
			args: args{first: math.SmallestNonzeroFloat64, second: 0.5},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Multiply(tt.args.first, tt.args.second)
			if (err != nil) != tt.wantErr {
				t.Errorf("Multiply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if actual != tt.want {
				t.Errorf("Multiply() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestCalculatorServiceImpl_Division(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		first  float64
		second float64
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "should divide the first operand by the second",
			args: args{first: 9, second: 3},
			want: 3,
		},
		{
			// Argument order matters: the reverse would give 3.
			name: "should respect the argument order",
			args: args{first: 3, second: 9},
			want: 1.0 / 3.0,
		},
		{
			name: "should return the first operand when dividing by one",
			args: args{first: 42, second: 1},
			want: 42,
		},
		{
			name: "should return zero when the first operand is zero",
			args: args{first: 0, second: 5},
			want: 0,
		},
		{
			name: "should produce a negative for mixed signs",
			args: args{first: -8, second: 2},
			want: -4,
		},
		{
			name: "should produce a positive for two negatives",
			args: args{first: -8, second: -2},
			want: 4,
		},
		{
			name: "should divide by a decimal",
			args: args{first: 1, second: 0.5},
			want: 2,
		},
		{
			name:    "should reject dividing by zero",
			args:    args{first: 8, second: 0},
			wantErr: true,
		},
		{
			name:    "should reject dividing zero by zero",
			args:    args{first: 0, second: 0},
			wantErr: true,
		},
		{
			name:    "should reject dividing by negative zero",
			args:    args{first: 8, second: math.Copysign(0, -1)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Division(tt.args.first, tt.args.second)
			if (err != nil) != tt.wantErr {
				t.Errorf("Division() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// The handler maps this sentinel to a 400, so identity matters.
				if !errors.Is(err, ErrDivideByZero) {
					t.Errorf("Division() error = %v, want ErrDivideByZero", err)
				}
				if actual != 0 {
					t.Errorf("Division() got = %v, want 0 alongside the error", actual)
				}
				return
			}
			if actual != tt.want {
				t.Errorf("Division() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

// Dividing by a small enough number leaves the representable range just as
// multiplying by a large one does. The handler rejects the result with a 400.
func TestCalculatorServiceImpl_Division_ShouldOverflow(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		first  float64
		second float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "should overflow to positive infinity when halving the largest float",
			args: args{first: math.MaxFloat64, second: 0.5},
			want: math.Inf(1),
		},
		{
			name: "should overflow when the divisor is tiny",
			args: args{first: 1e308, second: 1e-308},
			want: math.Inf(1),
		},
		{
			name: "should overflow to negative infinity for a negative dividend",
			args: args{first: -math.MaxFloat64, second: 0.5},
			want: math.Inf(-1),
		},
		{
			name: "should overflow to negative infinity for a negative divisor",
			args: args{first: math.MaxFloat64, second: -0.5},
			want: math.Inf(-1),
		},
		{
			name: "should stay finite just inside the range",
			args: args{first: math.MaxFloat64, second: 2},
			want: math.MaxFloat64 / 2,
		},
		{
			name: "should underflow to zero below the smallest subnormal",
			args: args{first: math.SmallestNonzeroFloat64, second: 2},
			want: 0,
		},
		{
			name: "should stay non-zero as a subnormal",
			args: args{first: 1e-308, second: 1e10},
			want: 1e-318,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Division(tt.args.first, tt.args.second)
			if err != nil {
				t.Errorf("Division() unexpected error = %v", err)
				return
			}
			if actual != tt.want {
				t.Errorf("Division() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

// The unary operations only ever shrink their operand, so no input can drive
// them out of range. This guards that property against a future change.
func TestCalculatorServiceImpl_UnaryOperations_ShouldNotOverflow(t *testing.T) {
	service := NewCalculatorService()
	inputs := []float64{
		math.MaxFloat64,
		math.MaxFloat64 / 2,
		1e308,
		1,
		0,
		math.SmallestNonzeroFloat64,
	}

	for _, value := range inputs {
		t.Run("should keep the square root finite", func(t *testing.T) {
			actual, err := service.SquareRoot(value)
			if err != nil {
				t.Errorf("SquareRoot(%v) unexpected error = %v", value, err)
				return
			}
			if math.IsInf(actual, 0) || math.IsNaN(actual) {
				t.Errorf("SquareRoot(%v) got = %v, want a finite value", value, actual)
			}
		})

		t.Run("should keep the percentage finite", func(t *testing.T) {
			actual, err := service.Percentage(value)
			if err != nil {
				t.Errorf("Percentage(%v) unexpected error = %v", value, err)
				return
			}
			if math.IsInf(actual, 0) || math.IsNaN(actual) {
				t.Errorf("Percentage(%v) got = %v, want a finite value", value, actual)
			}
		})
	}
}

func TestCalculatorServiceImpl_Exponentiation(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		base     float64
		exponent float64
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "should raise a positive base to a positive exponent",
			args: args{base: 2, exponent: 10},
			want: 1024,
		},
		{
			name: "should return one for a zero exponent",
			args: args{base: 7, exponent: 0},
			want: 1,
		},
		{
			name: "should return the base for an exponent of one",
			args: args{base: 7, exponent: 1},
			want: 7,
		},
		{
			name: "should invert for a negative exponent",
			args: args{base: 2, exponent: -2},
			want: 0.25,
		},
		{
			name: "should keep the sign for an odd exponent on a negative base",
			args: args{base: -2, exponent: 3},
			want: -8,
		},
		{
			name: "should drop the sign for an even exponent on a negative base",
			args: args{base: -2, exponent: 2},
			want: 4,
		},
		{
			name: "should treat a fractional exponent as a root",
			args: args{base: 9, exponent: 0.5},
			want: 3,
		},
		{
			name: "should return one for zero to the zero, as math.Pow does",
			args: args{base: 0, exponent: 0},
			want: 1,
		},
		{
			name: "should overflow to positive infinity",
			args: args{base: 10, exponent: 400},
			want: math.Inf(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Exponentiation(tt.args.base, tt.args.exponent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exponentiation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if actual != tt.want {
				t.Errorf("Exponentiation() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestCalculatorServiceImpl_Exponentiation_ShouldOverflow(t *testing.T) {
	service := NewCalculatorService()
	type args struct {
		base     float64
		exponent float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			// 2^1023 is the largest power of two that still fits.
			name: "should stay finite at the largest representable power of two",
			args: args{base: 2, exponent: 1023},
			want: math.Ldexp(1, 1023),
		},
		{
			name: "should overflow one power of two past the maximum",
			args: args{base: 2, exponent: 1024},
			want: math.Inf(1),
		},
		{
			name: "should overflow when squaring the largest float",
			args: args{base: math.MaxFloat64, exponent: 2},
			want: math.Inf(1),
		},
		{
			name: "should overflow to positive infinity for a large exponent",
			args: args{base: 10, exponent: 400},
			want: math.Inf(1),
		},
		{
			// An odd exponent keeps the sign of a negative base.
			name: "should overflow to negative infinity for an odd exponent",
			args: args{base: -10, exponent: 401},
			want: math.Inf(-1),
		},
		{
			name: "should overflow to positive infinity for an even exponent",
			args: args{base: -10, exponent: 400},
			want: math.Inf(1),
		},
		{
			name: "should overflow when dividing by zero through a negative exponent",
			args: args{base: 0, exponent: -1},
			want: math.Inf(1),
		},
		{
			name: "should keep the sign when negative zero has a negative odd exponent",
			args: args{base: math.Copysign(0, -1), exponent: -1},
			want: math.Inf(-1),
		},
		{
			name: "should underflow to zero for a large negative exponent",
			args: args{base: 10, exponent: -400},
			want: 0,
		},
		{
			name: "should underflow to zero below the smallest subnormal",
			args: args{base: 1e-200, exponent: 2},
			want: 0,
		},
		{
			// The smallest subnormal still survives as a non-zero value.
			name: "should stay non-zero at the smallest subnormal",
			args: args{base: 2, exponent: -1074},
			want: math.SmallestNonzeroFloat64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Exponentiation(tt.args.base, tt.args.exponent)
			if err != nil {
				t.Errorf("Exponentiation() unexpected error = %v", err)
				return
			}
			if actual != tt.want {
				t.Errorf("Exponentiation() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestCalculatorServiceImpl_Exponentiation_ShouldReturnNaN(t *testing.T) {
	service := NewCalculatorService()
	tests := []struct {
		name     string
		base     float64
		exponent float64
	}{
		{name: "should return NaN for a fractional power of a negative base", base: -2, exponent: 0.5},
		{name: "should return NaN for a cube root of a negative base", base: -8, exponent: 1.0 / 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Exponentiation(tt.base, tt.exponent)
			if err != nil {
				t.Errorf("Exponentiation() unexpected error = %v", err)
				return
			}
			if !math.IsNaN(actual) {
				t.Errorf("Exponentiation() got = %v, want NaN", actual)
			}
		})
	}
}

func TestCalculatorServiceImpl_SquareRoot(t *testing.T) {
	service := NewCalculatorService()
	tests := []struct {
		name    string
		value   float64
		want    float64
		wantErr bool
	}{
		{name: "should root a perfect square", value: 144, want: 12},
		{name: "should root zero", value: 0, want: 0},
		{name: "should root one", value: 1, want: 1},
		{name: "should root a decimal", value: 2.25, want: 1.5},
		{name: "should reject a negative value", value: -1, wantErr: true},
		{name: "should reject a negative decimal", value: -0.5, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.SquareRoot(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SquareRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// The handler maps this sentinel to a 400, so identity matters.
				if !errors.Is(err, ErrNegativeSquareRoot) {
					t.Errorf("SquareRoot() error = %v, want ErrNegativeSquareRoot", err)
				}
				return
			}
			if actual != tt.want {
				t.Errorf("SquareRoot() got = %v, want %v", actual, tt.want)
			}
		})
	}
}

func TestCalculatorServiceImpl_Percentage(t *testing.T) {
	service := NewCalculatorService()
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{name: "should divide a whole number by a hundred", value: 50, want: 0.5},
		{name: "should divide a hundred to one", value: 100, want: 1},
		{name: "should keep zero at zero", value: 0, want: 0},
		{name: "should keep the sign of a negative value", value: -25, want: -0.25},
		{name: "should divide a decimal", value: 12.5, want: 0.125},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := service.Percentage(tt.value)
			if err != nil {
				t.Errorf("Percentage() unexpected error = %v", err)
				return
			}
			if actual != tt.want {
				t.Errorf("Percentage() got = %v, want %v", actual, tt.want)
			}
		})
	}
}
