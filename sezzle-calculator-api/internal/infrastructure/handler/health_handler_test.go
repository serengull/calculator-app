package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sezzleerror "sezzle-calculator-api/internal/infrastructure/error"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// probe wires the health routes the way StartServer does and sends one request.
func probe(t *testing.T, method string, target string) *httptest.ResponseRecorder {
	t.Helper()

	e := echo.New()
	e.HTTPErrorHandler = sezzleerror.NewHTTPErrorHandler(false)
	e.Pre(middleware.RequestID())
	NewHealthHandler().Register(e)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(method, target, nil))

	return rec
}

func TestHealthHandler_LivenessHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{
			name:       "should report the process as alive",
			method:     http.MethodGet,
			target:     "/health/live",
			wantStatus: http.StatusOK,
		},
		{
			name:       "should not answer an unknown health route",
			method:     http.MethodGet,
			target:     "/health",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "should reject a write method",
			method:     http.MethodPost,
			target:     "/health/live",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := probe(t, tt.method, tt.target)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHealthHandler_LivenessHandler_ShouldReturnTheAliveStatus(t *testing.T) {
	rec := probe(t, http.MethodGet, "/health/live")

	var body HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not JSON: %v (body: %q)", err, rec.Body.String())
	}

	if body.Status != statusAlive {
		t.Errorf("status = %q, want %q", body.Status, statusAlive)
	}
}

// Liveness must stay independent of the calculator service: it answers even
// when nothing else is wired up, which is what makes it a process-level probe.
func TestHealthHandler_LivenessHandler_ShouldNotDependOnTheCalculatorService(t *testing.T) {
	e := echo.New()
	NewHealthHandler().Register(e)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
