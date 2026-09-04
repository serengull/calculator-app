package handler

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sezzle-calculator-api/internal/domain/service"
	sezzleerror "sezzle-calculator-api/internal/infrastructure/error"
	mock "sezzle-calculator-api/internal/mocks/service"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/mock/gomock"
)

type expectation func(recorder *mock.MockCalculatorServiceMockRecorder)

func request(t *testing.T, target string, expect expectation, exposeError bool) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	calculatorService := mock.NewMockCalculatorService(controller)
	if expect != nil {
		expect(calculatorService.EXPECT())
	}

	e := echo.New()
	e.HTTPErrorHandler = sezzleerror.NewHTTPErrorHandler(exposeError)
	e.Pre(middleware.RequestID())
	e.Use(middleware.Recover())
	NewCalculatorHandler(calculatorService).Register(e)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not JSON: %v (body: %q)", err, rec.Body.String())
	}

	return rec, body
}

func TestCalculatorHandler_CalculateHandler(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		expect      expectation
		wantStatus  int
		wantResult  float64
		wantMessage string
	}{
		{
			name:   "should route sum to the service",
			target: "/calculate?first=2&second=4&op=sum",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Sum(2.0, 4.0).Return(6.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 6,
		},
		{
			name:   "should route subtract to the service in argument order",
			target: "/calculate?first=2&second=4&op=subtract",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Subtract(2.0, 4.0).Return(-2.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: -2,
		},
		{
			name:   "should route multiply to the service",
			target: "/calculate?first=6&second=7&op=multiply",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Multiply(6.0, 7.0).Return(42.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 42,
		},
		{
			name:   "should route division to the service in argument order",
			target: "/calculate?first=9&second=3&op=division",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Division(9.0, 3.0).Return(3.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 3,
		},
		{
			name:   "should bind decimal operands",
			target: "/calculate?first=1.5&second=2.25&op=sum",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Sum(1.5, 2.25).Return(3.75, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 3.75,
		},
		{
			name:   "should bind negative operands",
			target: "/calculate?first=-8&second=-2&op=division",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Division(-8.0, -2.0).Return(4.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 4,
		},
		{
			name:   "should bind exponent notation",
			target: "/calculate?first=1e3&second=1&op=sum",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Sum(1000.0, 1.0).Return(1001.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 1001,
		},
		{
			name:        "should reject a missing first operand without calling the service",
			target:      "/calculate?op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "first is required",
		},
		{
			name:        "should reject a missing second operand for a binary operation",
			target:      "/calculate?first=2&op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "second is required",
		},
		{
			name:   "should accept an explicit zero operand",
			target: "/calculate?first=0&second=5&op=sum",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Sum(0.0, 5.0).Return(5.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 5,
		},
		{
			name:   "should route exponentiation to the service",
			target: "/calculate?first=2&second=10&op=exponentiation",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Exponentiation(2.0, 10.0).Return(1024.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 1024,
		},
		{
			// The unary operations take only first; second is bound but unused.
			name:   "should route sqrt to the service ignoring the second operand",
			target: "/calculate?first=144&second=99&op=sqrt",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.SquareRoot(144.0).Return(12.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 12,
		},
		{
			name:   "should route percentage to the service without a second operand",
			target: "/calculate?first=50&op=percentage",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Percentage(50.0).Return(0.5, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 0.5,
		},
		{
			name:   "should report a negative square root as a bad request",
			target: "/calculate?first=-1&op=sqrt",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.SquareRoot(-1.0).Return(0.0, service.ErrNegativeSquareRoot).Times(1)
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "cannot take the square root of a negative number",
		},
		{
			name:        "should reject a non-numeric first operand without calling the service",
			target:      "/calculate?first=abc&second=2&op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "first and second must be numbers",
		},
		{
			name:        "should reject a non-numeric second operand without calling the service",
			target:      "/calculate?first=2&second=abc&op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "first and second must be numbers",
		},
		{
			name:   "should report a divide-by-zero as a bad request",
			target: "/calculate?first=8&second=0&op=division",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Division(8.0, 0.0).Return(0.0, service.ErrDivideByZero).Times(1)
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "cannot divide by zero",
		},
		{
			// An error the handler does not recognise stays a 500.
			name:   "should report an unrecognised service error as an internal error",
			target: "/calculate?first=8&second=2&op=division",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Division(8.0, 2.0).Return(0.0, errors.New("disk on fire")).Times(1)
			},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal Server Error",
		},
		{
			name:        "should reject an infinite first operand",
			target:      "/calculate?first=Infinity&second=1&op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "first and second must be numbers",
		},
		{
			name:        "should reject a NaN operand",
			target:      "/calculate?first=NaN&second=1&op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "first and second must be numbers",
		},
		{
			name:        "should reject an infinite second operand",
			target:      "/calculate?first=1&second=-Inf&op=sum",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "first and second must be numbers",
		},
		{
			// A unary operation ignores second, so a junk value there is fine.
			name:   "should ignore an infinite second operand for a unary operation",
			target: "/calculate?first=9&second=Infinity&op=sqrt",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.SquareRoot(9.0).Return(3.0, nil).Times(1)
			},
			wantStatus: http.StatusOK,
			wantResult: 3,
		},
		{
			// encoding/json cannot marshal +Inf; without the guard this 500s.
			name:   "should reject a result that overflows to infinity",
			target: "/calculate?first=10&second=400&op=exponentiation",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Exponentiation(10.0, 400.0).Return(math.Inf(1), nil).Times(1)
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "result is out of range",
		},
		{
			name:   "should reject a NaN result",
			target: "/calculate?first=1&second=2&op=multiply",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Multiply(1.0, 2.0).Return(math.NaN(), nil).Times(1)
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "result is out of range",
		},
		{
			name:        "should reject an unknown operation without calling the service",
			target:      "/calculate?first=1&second=2&op=bogus",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "op must be one of: " + supportedOperations,
		},
		{
			name:        "should reject a missing operation without calling the service",
			target:      "/calculate?first=1&second=2",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "op must be one of: " + supportedOperations,
		},
		{
			name:        "should reject an operation in the wrong case",
			target:      "/calculate?first=1&second=2&op=SUM",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "op must be one of: " + supportedOperations,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, body := request(t, tt.target, tt.expect, false)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				result, ok := body["result"].(float64)
				if !ok {
					t.Fatalf("missing numeric Result field: %q", rec.Body.String())
				}
				if result != tt.wantResult {
					t.Errorf("Result = %v, want %v", result, tt.wantResult)
				}
				return
			}

			if body["message"] != tt.wantMessage {
				t.Errorf("message = %v, want %v", body["message"], tt.wantMessage)
			}
			if requestID, _ := body["requestId"].(string); requestID == "" {
				t.Error("error response has no requestId to correlate with the logs")
			}
		})
	}
}

func TestCalculatorHandler_CalculateHandler_ShouldPropagateServiceFailure(t *testing.T) {
	rec, body := request(t, "/calculate?first=1&second=2&op=sum", func(r *mock.MockCalculatorServiceMockRecorder) {
		r.Sum(1.0, 2.0).Return(0.0, errors.New("service exploded")).Times(1)
	}, false)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if body["message"] != "Internal Server Error" {
		t.Errorf("message = %v, want Internal Server Error", body["message"])
	}
	if strings.Contains(rec.Body.String(), "service exploded") {
		t.Error("the prod profile leaked the internal error text to the client")
	}
}

func TestCalculatorHandler_CalculateHandler_ShouldExposeInternalErrorOnlyInDev(t *testing.T) {
	_, body := request(t, "/calculate?first=1&second=2&op=sum", func(r *mock.MockCalculatorServiceMockRecorder) {
		r.Sum(1.0, 2.0).Return(0.0, errors.New("service exploded")).Times(1)
	}, true)

	if body["error"] != "service exploded" {
		t.Errorf("error = %v, want the raw internal error", body["error"])
	}
}

// The out-of-range guard sits after dispatch, so it must hold for every
// operation rather than only the one that overflows most easily.
func TestCalculatorHandler_CalculateHandler_ShouldRejectAnOutOfRangeResultForEveryOperation(t *testing.T) {
	tests := []struct {
		name   string
		target string
		expect expectation
	}{
		{
			name:   "sum",
			target: "/calculate?first=1&second=2&op=sum",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Sum(1.0, 2.0).Return(math.Inf(1), nil).Times(1)
			},
		},
		{
			name:   "subtract",
			target: "/calculate?first=1&second=2&op=subtract",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Subtract(1.0, 2.0).Return(math.Inf(-1), nil).Times(1)
			},
		},
		{
			name:   "multiply",
			target: "/calculate?first=1&second=2&op=multiply",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Multiply(1.0, 2.0).Return(math.Inf(1), nil).Times(1)
			},
		},
		{
			name:   "division",
			target: "/calculate?first=1&second=2&op=division",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Division(1.0, 2.0).Return(math.Inf(1), nil).Times(1)
			},
		},
		{
			name:   "exponentiation",
			target: "/calculate?first=1&second=2&op=exponentiation",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Exponentiation(1.0, 2.0).Return(math.Inf(1), nil).Times(1)
			},
		},
		{
			name:   "sqrt",
			target: "/calculate?first=1&op=sqrt",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.SquareRoot(1.0).Return(math.Inf(1), nil).Times(1)
			},
		},
		{
			name:   "percentage",
			target: "/calculate?first=1&op=percentage",
			expect: func(r *mock.MockCalculatorServiceMockRecorder) {
				r.Percentage(1.0).Return(math.NaN(), nil).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run("should reject an out-of-range "+tt.name+" result", func(t *testing.T) {
			rec, body := request(t, tt.target, tt.expect, false)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body: %q)", rec.Code, rec.Body.String())
			}
			if body["message"] != "result is out of range" {
				t.Errorf("message = %v, want result is out of range", body["message"])
			}
		})
	}
}

func TestCalculatorHandler_CalculateHandler_ShouldRecoverFromAPanickingService(t *testing.T) {
	rec, body := request(t, "/calculate?first=1&second=2&op=sum", func(r *mock.MockCalculatorServiceMockRecorder) {
		r.Sum(1.0, 2.0).DoAndReturn(func(float64, float64) (float64, error) {
			panic("service blew up")
		}).Times(1)
	}, false)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body: %q)", rec.Code, rec.Body.String())
	}
	if body["message"] != "Internal Server Error" {
		t.Errorf("message = %v, want Internal Server Error", body["message"])
	}
}
