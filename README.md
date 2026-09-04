# Sezzle Calculator

A calculator with a Go REST API and a React + TypeScript frontend, runnable
together with one command.

| Project | Stack | Detail |
|---|---|---|
| [`sezzle-calculator-api`](./sezzle-calculator-api) | Go 1.26, Echo v5 | [API README](./sezzle-calculator-api/README.md) |
| [`sezzle-calculator-web`](./sezzle-calculator-web) | React 19, TypeScript, Vite | [Web README](./sezzle-calculator-web/README.md) |

AI tooling was used throughout; the prompts are in [`PROMPTS.md`](./PROMPTS.md).

## Running with Docker

Requires Docker with Compose v2. Nothing else needs installing — Go and Node
are only used inside the build.

```sh
docker compose up --build
```

| Service | URL |
|---|---|
| Calculator UI | <http://localhost:3000> |
| API | <http://localhost:8080> |
| API docs (Swagger UI) | <http://localhost:8080/swagger/> |

nginx serves the built frontend and proxies `/api/*` to the API, so the browser
talks to a single origin and no CORS setup is needed. `docker compose down`
stops both.

Ports and the API profile can be overridden with a `.env` file: `WEB_PORT`,
`API_PORT`, `ACTIVE_PROFILE`, `LOG_LEVEL`, `VITE_API_BASE_URL` (a build
argument — changing it needs `docker compose build web`).

## Running without Docker

Requires Go 1.26 and Node 22. Two terminals:

```sh
# backend — http://localhost:8080
cd sezzle-calculator-api
go run ./cmd
```

```sh
# frontend — http://localhost:5173
cd sezzle-calculator-web
npm install
npm run dev
```

The Vite dev server proxies `/api/*` to `localhost:8080`, mirroring what nginx
does in Docker. The API must be run from its own directory, because the config
path is resolved relative to the working directory.

## API examples

One endpoint does the arithmetic:

```
GET /calculate?first=<number>&second=<number>&op=<operation>
```

```sh
$ curl 'http://localhost:8080/calculate?first=4&second=2&op=division'
{"result":2}

$ curl 'http://localhost:8080/calculate?first=2&second=10&op=exponentiation'
{"result":1024}

$ curl 'http://localhost:8080/calculate?first=144&op=sqrt'
{"result":12}
```

| `op` | Operation | Operands |
|---|---|---|
| `sum` | `first + second` | both |
| `subtract` | `first - second` | both |
| `multiply` | `first × second` | both |
| `division` | `first ÷ second` | both |
| `exponentiation` | `first ^ second` | both |
| `sqrt` | `√first` | `first` only |
| `percentage` | `first ÷ 100` | `first` only |

Errors share one body shape. `requestId` correlates with the server logs:

```sh
$ curl 'http://localhost:8080/calculate?first=4&second=0&op=division'
{"message":"cannot divide by zero","requestId":"zTBjRPVlfGxZNbjEXjgl"}
```

| Request | Status | `message` |
|---|---|---|
| `?first=x&second=2&op=sum` | 400 | `first and second must be numbers` |
| `?op=sum` | 400 | `first is required` |
| `?first=4&second=2&op=modulo` | 400 | `op must be one of: …` |
| `?first=4&second=0&op=division` | 400 | `cannot divide by zero` |
| `?first=-1&op=sqrt` | 400 | `cannot take the square root of a negative number` |
| `?first=10&second=400&op=exponentiation` | 400 | `result is out of range` |

A liveness probe backs the container healthcheck:

```sh
$ curl 'http://localhost:8080/health/live'
{"status":"ok"}
```

## Tests and coverage

```sh
cd sezzle-calculator-api && go test ./... && make cover
cd sezzle-calculator-web && npm test && npm run coverage
```

| | Tests | Coverage |
|---|---|---|
| API | 153 | 90.0% statements |
| Web | 105 | 97.7% statements, 90.8% branch |

Exported HTML reports are in [`coverage/`](./coverage) — `coverage/api/index.html`
and `coverage/web/index.html`.

The API figure excludes generated code (mocks, the Swagger spec) and process
wiring that is exercised by running the app rather than by unit tests
(`internal/server`, `cmd/`, the config manager). The exclusion list lives in the API
`Makefile`, and `coverage.full.out` keeps the unfiltered profile.

## Design decisions

**Arithmetic lives on the server.** The keypad sends every calculation to the
API rather than computing locally, so the API is the single source of truth.
The frontend owns only presentation: formatting, the pending-operation state
machine, and the keypad. The one exception is `+/−`, a sign flip on the current
entry rather than an arithmetic operation.

**Layering in the API.** `domain/service` holds arithmetic and knows nothing
about HTTP. `infrastructure/handler` owns binding, validation, and the
operation enum — the enum is a wire value, not a domain concept. The service
exports sentinel errors (`ErrDivideByZero`, `ErrNegativeSquareRoot`) that the
handler maps to 400; anything else stays a 500.

**One error shape, written in one place.** Handlers return `error` and never
write error responses. A single `HTTPErrorHandler` derives the status from the
error, writes the only error body, and gates internal detail behind the `dev`
profile. The UI renders the server's `message` verbatim rather than
reinterpreting it.

**Validation checks the raw query.** Echo's binder cannot distinguish an absent
parameter from an explicit zero, so presence is checked on the query string
itself. That keeps `?first=0&second=5&op=sum` valid while rejecting `?op=sum`,
which would otherwise quietly compute `0 + 0`.

**Non-finite values are rejected, not encoded.** `encoding/json` cannot
represent `±Inf` or `NaN`, and `ParseFloat` happily accepts `"Infinity"` from a
query string. Both operands and results are checked, so an overflow becomes a
400 rather than failing when the response is written.

**Same-origin in production.** nginx proxies `/api` to the API container, so the
`prod` profile needs no CORS origins at all. The `dev` profile allows `:5173`
for the Vite dev server.

**No UI component library.** The brief for the interface was a macOS-style
calculator, a pixel-specific target. A component library would have meant
overriding its design language on nearly every element, so the CSS is
hand-written. The tradeoff is one 436-line stylesheet.

## Assumptions

- `op` is case-sensitive; `SUM` is rejected rather than coerced.
- `percentage` means "divide by 100", matching the calculator's `%` key, rather
  than "x percent of y".
- Results carry float64 semantics — `0.1 + 0.2` returns `0.30000000000000004`.
  The API does not round; the UI trims to 12 significant digits for display and
  switches to exponential notation beyond 21 characters.

## Scope

Core to the brief: the four arithmetic operations, the REST API, the frontend,
unit tests with a coverage report, and the Docker setup.

Added beyond it: three more operations (exponentiation, square root,
percentage), a liveness endpoint, security response headers with a strict CSP on
the web container, and a display that shrinks long values to fit.

## Known limitations

- `?first=4&second=&op=sum` returns `200 {"result":4}` — an empty operand binds
  to zero instead of being rejected.
- The configuration package has no unit tests.
- The display is a single `aria-live` region, so a screen reader re-announces it
  on every keystroke rather than only on results.
