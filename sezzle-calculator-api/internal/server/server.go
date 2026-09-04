package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sezzle-calculator-api/internal/domain/service"
	"sezzle-calculator-api/internal/infrastructure/configuration"
	"sezzle-calculator-api/internal/infrastructure/docs"
	sezzleerror "sezzle-calculator-api/internal/infrastructure/error"
	"sezzle-calculator-api/internal/infrastructure/handler"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	devProfile      = "dev"
	gracefulTimeout = 5 * time.Second
)

func StartServer() {
	profile := os.Getenv("ACTIVE_PROFILE")
	if profile == "" {
		profile = devProfile
	}
	initLogger()

	config := configuration.NewConfigurationManager("./internal/infrastructure/resource", "application")

	e := echo.New()
	e.HTTPErrorHandler = sezzleerror.NewHTTPErrorHandler(profile == devProfile)

	e.Pre(middleware.RequestID())
	e.Use(middleware.Recover())
	if origins := config.GetServerConfig().AllowedOrigins; len(origins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: origins,
			AllowMethods: []string{http.MethodGet, http.MethodOptions},
		}))
	}

	calculatorHandler := handler.NewCalculatorHandler(service.NewCalculatorService())
	calculatorHandler.Register(e)

	healthHandler := handler.NewHealthHandler()
	healthHandler.Register(e)

	docs.RegisterRoutes(e)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := config.GetServerConfig().Port
	startConfig := echo.StartConfig{
		Address:         addr,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: gracefulTimeout,
		OnShutdownError: func(err error) {
			log.Error().Err(err).Dur("timeout", gracefulTimeout).Msg("graceful shutdown timed out")
		},
	}

	log.Info().Str("addr", addr).Msg("starting server")
	if err := startConfig.Start(ctx, e); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
	log.Info().Msg("server stopped gracefully")
}

func initLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	level, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	log.Logger = log.Logger.Level(level)

}
