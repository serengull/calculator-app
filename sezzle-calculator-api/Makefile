MOCKGEN     := go run go.uber.org/mock/mockgen@v0.6.0
SWAG        := go run github.com/swaggo/swag/cmd/swag@v1.16.4
SWAGGER_OUT := internal/infrastructure/docs/swagger
MOCK_SOURCE := internal/domain/service/calculator_service.go
MOCK_DEST   := internal/mocks/service/calculator_service_mock.go

# Generated code and process wiring that is exercised by running the app, not
# by unit tests: excluded so the figure reflects code under test.
COVER_EXCLUDE := /internal/mocks/|/internal/infrastructure/docs/|/internal/infrastructure/configuration/|/internal/server/|/cmd/

.PHONY: swagger
swagger:
	$(SWAG) init -g cmd/main.go -o $(SWAGGER_OUT) --packageName swagger --parseDependency --parseInternal
	go build ./internal/infrastructure/docs/...
	go test ./internal/infrastructure/docs/...

.PHONY: mockgen
mockgen:
	$(MOCKGEN) -source=$(MOCK_SOURCE) -destination=$(MOCK_DEST) -package=mock

.PHONY: cover
cover:
	go test ./... -count=1 -coverpkg=./... -coverprofile=coverage.full.out -covermode=atomic
	grep -vE '$(COVER_EXCLUDE)' coverage.full.out > coverage.out
	go tool cover -func=coverage.out | tail -1

.PHONY: cover-html
cover-html: cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "wrote coverage.html"
