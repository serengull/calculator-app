# Sezzle Calculator API

A small Go REST API that performs one arithmetic operation per request. It ships
alongside a React frontend in `../sezzle-calculator-web`.

## Run the whole project

From the **repository root** (the directory containing both projects):

```sh
docker compose up --build
```

That builds and starts both services:

| Service | URL | What it is |
|---|---|---|
| `web` | <http://localhost:3000> | The calculator UI, served by nginx |
| `api` | <http://localhost:8080> | This API (redirects to Swagger UI) |

nginx proxies `/api/*` to the API container with the prefix stripped, so the
browser talks to a single origin and no CORS setup is needed.

Ports and profile can be overridden with a `.env` file at the repository root:
`WEB_PORT`, `API_PORT`, `ACTIVE_PROFILE`, `LOG_LEVEL`, `VITE_API_BASE_URL`.

```sh
docker compose down          # stop
docker compose logs -f api   # follow logs
```

### API only

```sh
docker build -t sezzle-calculator-api .
docker run --rm -p 8080:8080 sezzle-calculator-api
```

The image compiles a static binary and runs it on Alpine as a non-root user
under the `prod` profile.

## Endpoints

### Calculate

```
GET /calculate?first=<number>&second=<number>&op=<operation>
```

`op` is case-sensitive and must be one of:

| `op` | Operation | Operands |
|---|---|---|
| `sum` | `first + second` | both |
| `subtract` | `first - second` | both |
| `multiply` | `first × second` | both |
| `division` | `first ÷ second` | both |
| `exponentiation` | `first ^ second` | both |
| `sqrt` | `√first` | `first` only |
| `percentage` | `first ÷ 100` | `first` only |

```sh
$ curl 'http://localhost:8080/calculate?first=4&second=2&op=division'
{"result":2}
```

### Liveness

```sh
$ curl 'http://localhost:8080/health/live'
{"status":"ok"}
```

Reports that the process is up. It checks no dependencies, so a failure means
the server itself is gone. Docker Compose uses it as the API's healthcheck, and
the `web` container waits for it before starting.

### Docs

Swagger UI is served by the API itself; the root redirects to it.

```
/                  302 -> /swagger/
/swagger/          Swagger UI
/swagger/doc.json  the generated OpenAPI spec
```

The spec is generated from annotations and embedded in the binary, so it works
from any working directory. Regenerate it with `make swagger`. Swagger UI's own
assets load from a CDN, so that page needs network access.

## Errors

Every error shares one body shape:

```json
{ "message": "cannot divide by zero", "requestId": "zTBjRPVlfGxZNbjEXjgl" }
```

`requestId` correlates with the server logs. An `error` field carrying the raw
internal error is added only under the `dev` profile.

| Request | Status | `message` |
|---|---|---|
| `?first=x&second=2&op=sum` | 400 | `first and second must be numbers` |
| `?op=sum` | 400 | `first is required` |
| `?first=4&op=sum` | 400 | `second is required` |
| `?first=4&second=2&op=modulo` | 400 | `op must be one of: …` |
| `?first=4&second=2&op=SUM` | 400 | `op must be one of: …` |
| `?first=4&second=0&op=division` | 400 | `cannot divide by zero` |
| `?first=-1&op=sqrt` | 400 | `cannot take the square root of a negative number` |
| `?first=Infinity&second=1&op=sum` | 400 | `first and second must be numbers` |
| `?first=10&second=400&op=exponentiation` | 400 | `result is out of range` |

Unary operations (`sqrt`, `percentage`) do not require `second`.

Operands and results must be finite. `Inf` and `NaN` parse as float64 but cannot
be represented in JSON, so both are rejected with a 400 rather than failing when
the response is written.

## Local development

Prerequisites: Go 1.26, Node 22.

**API** — serves on `:8080` under the default `dev` profile:

```sh
go run ./cmd
```

Select a profile with `ACTIVE_PROFILE` (`dev` | `test` | `prod`). An unknown
profile fails fast with a readable message.

```sh
ACTIVE_PROFILE=test go run ./cmd   # :8081
```

**Frontend** — serves on `:5173` and proxies `/api/*` to `:8080`:

```sh
cd ../sezzle-calculator-web
npm install
npm run dev
```

Run the API from this directory: the config path is resolved relative to the
working directory.

## Configuration

`internal/infrastructure/resource/application.yaml`, keyed by profile:

| Key | Meaning |
|---|---|
| `server.port` | Listen address, including the leading colon (`:8080`) |
| `server.allowedOrigins` | CORS origins. Empty disables CORS entirely |
| `server.staticDir` | Built frontend assets to serve at `/`. Empty disables it |

In Docker the frontend is served by nginx on the same origin, so `prod` needs no
CORS origins. The `dev` profile allows `:5173` for `npm run dev`.

## Tests

```sh
go test ./...
make cover        # coverage summary
make cover-html   # browsable report at coverage.html
```

`make cover` uses `-coverpkg=./...`, because by default Go instruments only the
package under test and under-reports shared code such as the error handler.

It then excludes generated code and process wiring — `internal/mocks`,
`internal/infrastructure/docs`, `internal/infrastructure/configuration`,
`internal/server` and `cmd/` — so the figure reflects code that unit tests are meant
to cover. `coverage.full.out` keeps the unfiltered profile if you need it.

## Codegen

```sh
make swagger    # regenerate the OpenAPI spec from annotations
make mockgen    # regenerate the gomock double for CalculatorService
```

Run `make mockgen` after changing the `CalculatorService` interface.

## Layout

```
cmd/                          entrypoint
internal/
  domain/service/             arithmetic, no HTTP knowledge
  infrastructure/
    handler/                  HTTP layer: binding, op enum, validation
    error/                    SezzleError + global error handler
    configuration/            viper profile config
    resource/                 application.yaml
    docs/                     Swagger UI + generated spec
  server/                     wiring, middleware, graceful shutdown
  mocks/service/              generated gomock double
Dockerfile                    API image
```

## Design notes

**Layering.** `domain/service` holds arithmetic and knows nothing about HTTP.
The operation enum lives in the handler because it is a wire value, not a domain
concept. Unary operations take a single argument in the service; the handler
wraps them so all operations share one dispatch signature.

**Errors.** Handlers return `error` and never write responses themselves.
`SezzleError` carries a status and implements `echo.HTTPStatusCoder`, so a
single error handler derives the status, writes the only error body, and gates
internal detail behind the `dev` profile. The service exports sentinel errors
(`ErrDivideByZero`, `ErrNegativeSquareRoot`) that the handler maps to 400 —
anything else stays a 500.

**Validation.** Echo's binder cannot distinguish an absent query parameter from
an explicit zero, so the handler checks the raw query for presence. That keeps
`?first=0&second=5&op=sum` valid while rejecting `?op=sum`.

**Graceful shutdown.** Echo v5 removed `e.Shutdown`; shutdown is driven by
cancelling the context passed to `StartConfig.Start`. `signal.NotifyContext`
triggers it on `SIGINT`/`SIGTERM`, with a 5s window for in-flight requests.
