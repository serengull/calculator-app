package main

import (
	"sezzle-calculator-api/internal/server"
)

// @title           Sezzle Calculator API
// @version         1.0.0
// @description     A single-endpoint arithmetic API.
// @description     Every error response shares one shape. requestId correlates with the server logs, and error carries the raw internal error but is only populated under the dev profile.
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /
// @schemes         http https
func main() {
	server.StartServer()
}
