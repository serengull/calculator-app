# Sezzle Calculator — Web Frontend

A small React + TypeScript single-page app for the Go calculator API in the parent
directory. Built with Vite, tested with Vitest + React Testing Library, styled with
hand-written CSS (no UI framework).

## Requirements

- Node.js 20.19+ or 22.12+ (developed on Node 26)
- The Go backend running locally (`dev` profile → port 8080, `test` profile → 8081)

## Setup

```sh
cd web
npm install
npm run dev          # http://localhost:5173
```

## Scripts

| Script | What it does |
| --- | --- |
| `npm run dev` | Vite dev server with the `/api` proxy |
| `npm run build` | Type-checks (`tsc -b`) then produces a production bundle in `dist/` |
| `npm run preview` | Serves the built bundle |
| `npm test` | Runs the test suite once |
| `npm run test:watch` | Watch mode |
| `npm run coverage` | `vitest run --coverage` — text summary plus `coverage/index.html` and `lcov.info` |
| `npm run typecheck` | `tsc -b --force` — type-checks both TS projects (app + vite config) |

## API contract

Single endpoint:

```
GET /calculate?first=<number>&second=<number>&op=<sum|subtract|multiply|division>
```

- `200` → `{"result": 6}`
- `400` / `5xx` → `{"message": "...", "requestId": "...", "error": "..."}`
  (`error` only appears under the backend's `dev` profile)

The three backend error messages (`first and second must be numbers`,
`op must be one of: sum, subtract, multiply, division`, `cannot divide by zero`) are
rendered verbatim — the client never rewrites or swallows a server `message`.

## Dev proxy and environment variables

`vite.config.ts` proxies `/api/*` to the Go backend and strips the `/api` prefix, so
a browser request to `/api/calculate` reaches `http://localhost:8080/calculate`. This
means development works even if backend CORS is unavailable.

Copy `.env.example` to `.env` (or `.env.local`) to override:

| Variable | Default | Purpose |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `/api` | Base URL the API client prefixes to `/calculate`. `/api` is a **development convenience, not a requirement** — it exists to match the dev-server proxy below. Set it to `/` for same-origin production builds, or to an absolute origin such as `http://localhost:8081` to hit the `test` profile directly (requires backend CORS). |
| `VITE_PROXY_TARGET` | `http://localhost:8080` | Where the Vite dev server forwards `/api`. Dev-server only; has no effect on a production build. |

`VITE_API_BASE_URL` is read at **build time** (Vite inlines it), so it must be set
before `npm run build`, not at container start.

### Production / Docker: same origin, no proxy

The Docker image serves the built frontend as static files at `/` and the API at
`/calculate` from the **same Go server, on the same origin**. There is no `/api`
prefix and no proxy in that setup. The image therefore sets:

```
VITE_API_BASE_URL=/
```

`resolveBaseUrl()` strips the trailing slash, yielding an empty base, so the client
requests `/calculate?...` directly against its own origin — no CORS involved. This
path is verified end to end in the image.

Summary of the two supported deployments:

| Deployment | `VITE_API_BASE_URL` | How the request reaches Go |
| --- | --- | --- |
| `npm run dev` | `/api` (default) | Vite proxy strips `/api`, forwards to `VITE_PROXY_TARGET` |
| Docker / production | `/` | Same-origin request to `/calculate`, served by the Go binary |

A third option — hosting `dist/` separately from the backend — works by setting
`VITE_API_BASE_URL` to the backend's absolute origin, which requires the backend's
`middleware.CORS()` to allow that origin.

## Structure and rationale

```
src/
  api/
    types.ts          Operation enum, request/response shapes, display labels
    ApiError.ts       ApiError (HTTP failure w/ message, status, requestId, detail)
                      and NetworkError (transport failure)
    client.ts         calculate() — the only module that touches fetch
  lib/
    validation.ts     Pure form validation, no React, no network
  components/
    CalculatorForm.tsx  Form state, submit orchestration, error mapping
    ResultPanel.tsx     Presentational result / error / loading / idle states
  App.tsx             Page shell
  styles.css          All styling
```

Design decisions:

- **Transport isolated in `api/client.ts`.** Every network detail — query-string
  construction, response parsing, error classification — lives in one place, so
  components deal in typed values and typed errors rather than `Response` objects.
  Nothing in the API layer uses `any`; untrusted JSON is narrowed through
  `unknown` + type guards.
- **Two error classes instead of error codes.** `ApiError` carries the server's
  `message`, `status`, `requestId` and (dev-profile) `error`; `NetworkError` marks a
  request that never got a response. `instanceof` checks in the component keep the
  branching obvious.
- **Validation is a pure function.** `validate()` takes raw string form values and
  returns either typed values or per-field messages. Being React-free makes it
  cheap to test exhaustively and reusable if a second entry point ever appears.
- **Two validation layers, server wins.** Client-side checks (empty, non-numeric,
  divide-by-zero) exist to avoid pointless round trips and give per-field feedback.
  They never replace the server's answer: any `message` the backend returns is shown
  as-is, along with the `requestId` for support/log correlation.
- **Inputs are `type="text"` with `inputMode="decimal"`,** not `type="number"`.
  Number inputs silently discard invalid text in most browsers, which would make the
  "enter a valid number" path untestable and unreachable; text inputs let the app own
  validation while `inputMode` still brings up a numeric keypad on mobile.
- **In-flight requests are aborted** when a new submit or a reset happens, so a slow
  earlier response cannot overwrite a newer result.
- **Local state only.** One form and one result do not justify a store or a data
  fetching library.
- **Responsive by layout, not breakpoint soup.** A single grid collapses from
  three columns to one under 560px; controls go full-width. A `prefers-color-scheme`
  block supplies a dark palette via the same CSS custom properties.

## Testing

Vitest with the jsdom environment; `fetch` is stubbed per-test with `vi.stubGlobal`
(no MSW). Coverage is collected with `@vitest/coverage-v8`.

- `src/lib/validation.test.ts` — number parsing and every validation rule
- `src/api/client.test.ts` — base-URL resolution, query building, each backend error
  message, a 5xx, an unparseable body, a malformed 200, and a network failure
- `src/components/CalculatorForm.test.tsx` — submit → result, each client-side block,
  server error rendering, network error rendering, reset, result formatting
- `src/App.test.tsx` — shell smoke test

```sh
npm run coverage
```
