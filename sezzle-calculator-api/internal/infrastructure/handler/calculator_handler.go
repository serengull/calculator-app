package handler

import (
	"errors"
	"math"
	"net/http"
	"sezzle-calculator-api/internal/domain/service"
	sezzleerror "sezzle-calculator-api/internal/infrastructure/error"

	"github.com/labstack/echo/v5"
)

type CalculatorHandler struct {
	calculatorService service.CalculatorService
}

type OperationKind string

const supportedOperations = "sum, subtract, multiply, division, exponentiation, sqrt, percentage"

const (
	Sum      OperationKind = "sum"
	Subtract OperationKind = "subtract"
	Multiply OperationKind = "multiply"
	Division OperationKind = "division"

	Exponentiation OperationKind = "exponentiation"
	SquareRoot     OperationKind = "sqrt"
	Percentage     OperationKind = "percentage"
)

type CalculateRequest struct {
	First         float64       `query:"first"`
	Second        float64       `query:"second"`
	OperationKind OperationKind `query:"op"`
}

type CalculateResponse struct {
	Result float64 `json:"result"`
}

func NewCalculatorHandler(calculatorService service.CalculatorService) *CalculatorHandler {
	return &CalculatorHandler{calculatorService: calculatorService}
}

// CalculateHandler godoc
// @Summary      Perform arithmetic operation
// @Tags         calculator
// @Produce      json
// @Param        first   query  number  true  "First operand"
// @Param        second  query  number  true  "Second operand"
// @Param        op      query  string  true  "Operation" Enums(sum, subtract, multiply, division, exponentiation, sqrt, percentage)
// @Success      200     {object} CalculateResponse
// @Failure      400     {object} sezzleerror.Response
// @Failure      500     {object} sezzleerror.Response
// @Router       /calculate [get]
func (ch *CalculatorHandler) CalculateHandler(c *echo.Context) error {
	var req CalculateRequest
	if err := c.Bind(&req); err != nil {
		return sezzleerror.BadRequest("first and second must be numbers")
	}

	op := ch.handleOp(req.OperationKind)
	if op == nil {
		return sezzleerror.BadRequest("op must be one of: " + supportedOperations)
	}

	if err := validateOperands(c, req); err != nil {
		return err
	}

	result, err := op(req.First, req.Second)
	if err != nil {
		if errors.Is(err, service.ErrDivideByZero) || errors.Is(err, service.ErrNegativeSquareRoot) {
			return sezzleerror.BadRequest(err.Error())
		}
		return err
	}

	if !isFinite(result) {
		return sezzleerror.BadRequest("result is out of range")
	}

	return c.JSON(http.StatusOK, CalculateResponse{Result: result})
}

// isFinite reports whether the value can be represented in JSON.
func isFinite(value float64) bool {
	return !math.IsInf(value, 0) && !math.IsNaN(value)
}

// isUnary reports whether the operation is computed from first alone.
func isUnary(op OperationKind) bool {
	return op == SquareRoot || op == Percentage
}

// validateOperands rejects absent operands. Binding cannot tell a missing
// query parameter from an explicit zero, so the raw query is checked instead.
func validateOperands(c *echo.Context, req CalculateRequest) error {
	query := c.Request().URL.Query()
	binary := !isUnary(req.OperationKind)

	if !query.Has("first") {
		return sezzleerror.BadRequest("first is required")
	}
	if binary && !query.Has("second") {
		return sezzleerror.BadRequest("second is required")
	}

	// ParseFloat accepts "Inf" and "NaN", which are not usable operands.
	if !isFinite(req.First) || (binary && !isFinite(req.Second)) {
		return sezzleerror.BadRequest("first and second must be numbers")
	}

	return nil
}

func (ch *CalculatorHandler) handleOp(op OperationKind) func(first, second float64) (float64, error) {
	switch op {
	case Sum:
		return ch.calculatorService.Sum
	case Subtract:
		return ch.calculatorService.Subtract
	case Multiply:
		return ch.calculatorService.Multiply
	case Division:
		return ch.calculatorService.Division
	case Exponentiation:
		return ch.calculatorService.Exponentiation
	// The unary operations ignore second so they share one dispatch signature.
	case SquareRoot:
		return func(first, _ float64) (float64, error) {
			return ch.calculatorService.SquareRoot(first)
		}
	case Percentage:
		return func(first, _ float64) (float64, error) {
			return ch.calculatorService.Percentage(first)
		}
	default:
		return nil
	}
}

func (ch *CalculatorHandler) Register(e *echo.Echo) {
	e.GET("/calculate", ch.CalculateHandler)
}
