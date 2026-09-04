package error

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

const (
	resourceNotFoundTitle    = "Not found"
	resourceNotFoundMessage  = "Resource not found"
	resourceConflictTitle    = "Already exists"
	resourceConflictMessage  = "Resource already exists"
	internalServerErrorTitle = "Internal Server Error"
)

func (err SezzleError) Error() string {
	return err.Detail
}

func (err SezzleError) StatusCode() int {
	return err.Status
}

func (err SezzleError) MarshalZerologObject(e *zerolog.Event) {
	e.Int("status", err.Status).
		Str("title", err.Title).
		Str("detail", err.Detail).
		Str("message", err.Message).
		Str("request_method", err.RequestMethod).
		Str("request_uri", err.RequestUri)
}

func BadRequest(detail string) error {
	return makeSezzleError(http.StatusBadRequest, detail, "", "")
}

func InternalServerError(detail string) error {
	return makeSezzleError(http.StatusInternalServerError, detail, internalServerErrorTitle, "")
}

func NotFound(detail string) error {
	return makeSezzleError(http.StatusNotFound, detail, resourceNotFoundTitle, resourceNotFoundMessage)
}

func Conflict(detail string) error {
	return makeSezzleError(http.StatusConflict, detail, resourceConflictTitle, resourceConflictMessage)
}

func makeSezzleError(code int, detail string, title string, message string) error {
	return &SezzleError{
		Title:   title,
		Message: message,
		Status:  code,
		Detail:  detail,
		Instant: time.Now(),
	}
}
