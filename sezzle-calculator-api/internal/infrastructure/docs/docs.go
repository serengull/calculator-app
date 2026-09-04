package docs

import (
	_ "embed"
	"net/http"

	_ "sezzle-calculator-api/internal/infrastructure/docs/swagger"

	"github.com/labstack/echo/v5"
)

//go:embed swagger/swagger.json
var openAPISpec []byte

//go:embed swagger.html
var swaggerUI []byte

const (
	BasePath = "/swagger"
	SpecPath = BasePath + "/doc.json"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET(BasePath, func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, BasePath+"/")
	})

	ui := func(c *echo.Context) error {
		return c.HTMLBlob(http.StatusOK, swaggerUI)
	}
	e.GET(BasePath+"/", ui)
	e.GET(BasePath+"/index.html", ui)
	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusFound, BasePath+"/")
	})
	e.GET(SpecPath, func(c *echo.Context) error {
		return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, openAPISpec)
	})
}
