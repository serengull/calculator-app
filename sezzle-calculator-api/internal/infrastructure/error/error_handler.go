package error

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog/log"
)

type SezzleError struct {
	Title         string    `json:"title"`
	Status        int       `json:"status"`
	Detail        string    `json:"detail"`
	RequestUri    string    `json:"requestUri"`
	RequestMethod string    `json:"requestMethod"`
	Instant       time.Time `json:"instant"`
	Message       string    `json:"message"`
}

type Response struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
	Error     string `json:"error,omitempty"`
}

func NewHTTPErrorHandler(exposeError bool) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if err == nil {
			return
		}

		if res, _ := echo.UnwrapResponse(c.Response()); res != nil && res.Committed {
			return
		}

		code, message := statusAndMessage(err)
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)

		l := log.With().
			Int("status", code).
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Str("route", c.Path()).
			Str("ip", c.RealIP()).
			Str("request_id", requestID).
			Logger()

		if code >= http.StatusInternalServerError {
			l.Error().Err(err).Msg("request failed")
		} else {
			l.Warn().Err(err).Msg("request rejected")
		}

		body := Response{Message: message, RequestID: requestID}
		if exposeError {
			body.Error = err.Error()
		}

		var writeErr error
		if c.Request().Method == http.MethodHead {
			writeErr = c.NoContent(code)
		} else {
			writeErr = c.JSON(code, body)
		}
		if writeErr != nil {
			l.Error().Err(writeErr).Msg("failed to write error response")
		}
	}
}

func statusAndMessage(err error) (int, string) {
	code := http.StatusInternalServerError

	var coder echo.HTTPStatusCoder
	if errors.As(err, &coder) {
		if sc := coder.StatusCode(); sc != 0 {
			code = sc
		}
	}

	var sezzleErr *SezzleError
	if errors.As(err, &sezzleErr) {
		message := sezzleErr.Message
		if message == "" {
			message = sezzleErr.Detail
		}
		if message == "" {
			message = http.StatusText(code)
		}
		return code, message
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) && httpErr.Message != "" {
		return code, httpErr.Message
	}

	return code, http.StatusText(code)
}
