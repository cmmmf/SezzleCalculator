# Sezzle Calculator

A full-stack calculator: a **React + TypeScript** keypad UI backed by a **Go (Gin) REST API** that does
all the math. It supports add, subtract, multiply, divide, power, square root and percentage, validates
input on both sides, and reports every edge case (division by zero, overflow, undefined results, bad
input) with a stable error code and a friendly message.

- [Quick start (Docker)](#quick-start-docker)
- [Running locally](#running-locally)
- [API](#api)
- [Using the calculator](#using-the-calculator)
- [Tests and coverage](#tests-and-coverage)
- [Project structure](#project-structure)
- [Design decisions](#design-decisions)
- [Assumptions](#assumptions)
- [What I'd do with more time](#what-id-do-with-more-time)
- [AI usage](#ai-usage)

---

## Quick start (Docker)

Requires Docker with Compose v2.

```bash
docker compose up --build
```

| URL | What |
|---|---|
| http://localhost:8080 | Calculator UI (nginx, which also proxies `/api` to the backend) |
| http://localhost:8080/swagger | Interactive API docs (Swagger UI) |
| http://localhost:5000 | Backend API, exposed directly |

If a port is taken, override it: `FRONTEND_PORT=3000 BACKEND_PORT=5001 docker compose up --build`
(PowerShell: `$env:FRONTEND_PORT="3000"; docker compose up --build`).

Stop with `Ctrl+C`, then `docker compose down`.

---

## Running locally

Prerequisites: **Go 1.26+** and **Node.js 24+** (npm 11).

### Backend

```bash
cd backend
go run ./cmd/server          # http://localhost:8080
```

Configuration comes from environment variables (all optional):

| Variable | Default | Values |
|---|---|---|
| `PORT` | `8080` | 1–65535 |
| `GIN_MODE` | `release` | `debug`, `release`, `test` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json`, `console` (readable, for local work) |
| `CORS_ORIGINS` | `http://localhost:5173` | comma-separated origins, or `*` |

```bash
LOG_FORMAT=console LOG_LEVEL=debug go run ./cmd/server
```

### Frontend

In a second terminal, with the backend running:

```bash
cd frontend
npm ci
npm run dev                  # http://localhost:5173
```

The Vite dev server proxies `/api`, `/healthz` and `/swagger` to `http://localhost:8080`, so the
browser only ever talks to one origin. Point it elsewhere with `VITE_BACKEND_URL=http://localhost:5000 npm run dev`.

Other scripts: `npm run build`, `npm run lint` (oxlint), `npm run typecheck`, `npm test`, `npm run coverage`.

---

## API

Base path `/api/v1`. Every operation is a `POST` with a JSON body. The full contract (schemas, examples and
which errors each route can return) is in [`backend/internal/docs/openapi.yaml`](backend/internal/docs/openapi.yaml)
and is browsable at `/swagger` on a running server.

| Method | Path | Body | Result |
|---|---|---|---|
| POST | `/api/v1/add` | `{"a": 10, "b": 4}` | `14` |
| POST | `/api/v1/subtract` | `{"a": 10, "b": 4}` | `6` |
| POST | `/api/v1/multiply` | `{"a": 10, "b": 4}` | `40` |
| POST | `/api/v1/divide` | `{"a": 10, "b": 4}` | `2.5` |
| POST | `/api/v1/power` | `{"a": 2, "b": 10}` | `1024` |
| POST | `/api/v1/percentage` | `{"a": 15, "b": 200}` | `30` (**a % of b**) |
| POST | `/api/v1/sqrt` | `{"a": 16}` | `4` |
| GET | `/healthz` | – | `{"status":"ok"}` |
| GET | `/swagger` | – | Swagger UI (`/swagger/openapi.yaml` is the raw spec) |

**Success** → `200 {"operation":"divide","result":2.5}`

**Error** → `{"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}}`.
Clients should branch on `code`; `message` is for humans and may change.

| Situation | Status | Code |
|---|---|---|
| Malformed JSON, missing/null field, wrong type, out-of-range number, unknown field, trailing data | 400 | `INVALID_REQUEST` |
| Divide by zero | 422 | `DIVISION_BY_ZERO` |
| Square root of a negative number | 422 | `NEGATIVE_SQUARE_ROOT` |
| Result is ±Inf (e.g. `power(10, 400)`, `power(0, -1)`) | 422 | `RESULT_OUT_OF_RANGE` |
| Result is NaN (e.g. `power(-8, 0.5)`) | 422 | `UNDEFINED_RESULT` |
| Unknown route | 404 | `NOT_FOUND` |
| Wrong HTTP method on a known route | 405 | `METHOD_NOT_ALLOWED` |
| Body larger than 1 KB | 413 | `PAYLOAD_TOO_LARGE` |
| Panic / unexpected error | 500 | `INTERNAL_ERROR` |

Every response carries an `X-Request-ID` header: the one the client sent, or a generated one. The same ID
is in the server's log line for that request.

### Examples

These use the local backend on port 8080; with Docker use `localhost:5000` (direct) or `localhost:8080`
(through nginx).

```bash
curl -s -X POST http://localhost:8080/api/v1/divide \
  -H 'Content-Type: application/json' -d '{"a": 10, "b": 4}'
# {"operation":"divide","result":2.5}

curl -s -X POST http://localhost:8080/api/v1/percentage \
  -H 'Content-Type: application/json' -d '{"a": 15, "b": 200}'
# {"operation":"percentage","result":30}

curl -s -X POST http://localhost:8080/api/v1/sqrt \
  -H 'Content-Type: application/json' -d '{"a": 16}'
# {"operation":"sqrt","result":4}

curl -s -i -X POST http://localhost:8080/api/v1/divide \
  -H 'Content-Type: application/json' -d '{"a": 1, "b": 0}'
# HTTP/1.1 422 Unprocessable Entity
# {"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}}

curl -s -X POST http://localhost:8080/api/v1/add \
  -H 'Content-Type: application/json' -d '{"a": 1}'
# {"error":{"code":"INVALID_REQUEST","message":"field \"b\" is required"}}

curl -s -X POST http://localhost:8080/api/v1/power \
  -H 'Content-Type: application/json' -d '{"a": 10, "b": 400}'
# {"error":{"code":"RESULT_OUT_OF_RANGE","message":"result is too large to be represented"}}
```

On Windows PowerShell, call `curl.exe` and escape the inner quotes, or use `Invoke-RestMethod`:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/add `
  -ContentType 'application/json' -Body '{"a": 10, "b": 4}'
```

---

## Using the calculator

Type the first number, pick an operator, type the second number and press `=`. The UI sends the
operation to the API and shows the result with the expression above it (`2 + 3 =`).

- **Chaining**: `2 + 3 ×` evaluates `2 + 3` first and continues with `5 ×`.
- **Percentage** is an operator: `15 % 200 =` gives `30` (15 % of 200), the same as the API.
- **Square root** applies to the number being typed, or to the last result. In the middle of an
  expression (`2 + 9 √`) it asks you to press `=` first instead of guessing.
- **±** flips the sign of the number being typed (or of the last result).
- **Errors** appear in a banner and keep your input, so after `5 ÷ 0 =` you can press `⌫`, type `2`
  and press `=` again.
- **Validation before calling the API**: one decimal point, at most 15 digits, and a second number
  is required before `=`.
- `=` is disabled while a request is in flight; `AC` cancels it.

| Key | Action |
|---|---|
| `0`–`9`, `.` or `,` | Digits, decimal point |
| `+` `-` `*` (or `x`) `/` `^` `%` | Operators |
| `r` | Square root |
| `n` | Toggle sign |
| `Enter` or `=` | Evaluate |
| `Backspace` | Delete last digit |
| `Esc` or `Delete` | Clear |

The layout is a CSS grid that fits 320 px wide screens and follows the system light/dark theme.

---

## Tests and coverage

### Backend

```bash
cd backend
go vet ./...
go test ./... -race -coverprofile=coverage.out
go tool cover -func=coverage.out              # per-function summary
go tool cover -html=coverage.out -o coverage.html
```

`-race` needs cgo (a C compiler). On Windows without gcc, drop `-race`.

`internal/` is at **99% statement coverage**. The `cmd/server` wiring is covered by running it (and by
the Docker health check), not by unit tests. The tests are table-driven with testify:

- **service**: every operation, including `-0` normalization, overflow, NaN and non-finite operands.
- **controller**: uses a fake `CalculatorService`, so it tests HTTP binding and error mapping only:
  each domain error to its status and code, every invalid-body case, and that unexpected errors
  never leak details.
- **middleware**: request ID, log level by status (with `zaptest/observer`), panic recovery, CORS and the
  body limit.
- **router**: end to end through the real service and middleware chain, 404/405/413, and a check that
  **the OpenAPI spec documents exactly the routes the server registers**, so the docs can't drift.
- **config** and **logger**: defaults and validation of every variable.

### Frontend

```bash
cd frontend
npm test                # Vitest
npm run coverage        # report in frontend/coverage/index.html
```

119 tests, **100% line coverage**, using Vitest and React Testing Library:

- **calculatorService**: request shape, every API error code mapped to its message, network failures,
  aborts and malformed responses, all with a mocked `fetch`.
- **calculatorReducer**: the whole keypad state machine as a pure function (input limits, chaining,
  sign, sqrt rules, errors, stale responses).
- **useCalculator**: sends pending requests, aborts on clear and unmount, ignores late responses.
- **User flows** (`Calculator.test.tsx`): the real client, hook and components against a fake backend,
  including `2 + 3 =` → `5`, `5 ÷ 0 =` → error banner and recovery, keyboard input, `=` disabled while
  loading, and network errors.

Coverage output is not committed; the commands above generate it.

---

## Project structure

```
backend/
├── cmd/server/main.go         wiring: config → logger → service → controller → router → HTTP server
└── internal/
    ├── config/                env vars → validated Config
    ├── logger/                zap logger from config
    ├── model/                 request/response DTOs and error codes (the wire contract)
    ├── service/               CalculatorService: pure math + domain errors (no HTTP, no logging)
    ├── controller/            Gin handlers: strict binding, call service, map errors → HTTP
    ├── middleware/            request ID, request logging, panic recovery, CORS, body limit
    ├── router/                Gin engine, middleware chain and routes
    └── docs/                  embedded OpenAPI spec + Swagger UI page
frontend/src/
├── services/calculatorService.ts   typed API client, error codes → user messages
├── state/calculatorReducer.ts      pure keypad state machine + keyboard mapping
├── state/formatNumber.ts           display-only number formatting
├── hooks/useCalculator.ts          runs the reducer's pending requests against the API
├── hooks/useKeyboardShortcuts.ts   page-wide keyboard input
└── components/                     Calculator, Display, Keypad, ErrorBanner (render only)
docker-compose.yml                  backend + nginx frontend
```

---

## Design decisions

**Layered backend with dependencies pointing one way.** router → controller → service. The service is
plain Go behind an interface (`CalculatorService`) and returns sentinel errors (`ErrDivisionByZero`, …).
The controller handles HTTP only and maps those errors to status codes through one table. Everything is
built in `main.go` and passed through constructors, with no globals (not even the logger). So each layer
can be tested on its own: controller tests use a fake service, and service tests never touch HTTP.

**One route per operation** instead of a single `/calculate` endpoint with an `op` field. Each operation
gets its own request schema (`sqrt` takes only `a`), its own documented errors, and readable logs and
URLs. Adding an operation means a service method, a one-line handler and a route.

**Strict input validation.** Operands are `*float64` with `binding:"required"`, so a missing `b` is an
error rather than a silent `0`. The decoder rejects unknown fields (`{"a":1,"b":2,"c":3}`), trailing data
and numbers outside float64 range, and the body is capped at 1 KB. Error messages say which field is wrong.

**400 vs 422.** 400 means the request itself is malformed. 422 means the request is valid, but the math
has no real, finite answer (division by zero, √ of a negative number, overflow, NaN). Clients can tell
"fix your request" apart from "this calculation has no answer".

**Numbers are float64 and the API never rounds.** The API returns exactly what IEEE-754 gives, so
`0.1 + 0.2` is `0.30000000000000004`. The frontend rounds for display only, to 15 significant digits
(the most a float64 always stores exactly), so it shows `0.3` while keeping full precision for the next
operation. The service turns `-0` into `0`, and NaN/±Inf become errors, so clients only ever get finite
JSON numbers. **For money I would not use float64**: amounts would be integer minor units (cents) or a
decimal type, with explicit rounding rules.

**The backend is the only place that does math.** The frontend never computes results (except flipping
a result's sign, which is exact). The keypad logic is a **pure reducer**: pressing `=` doesn't fetch, it
records a `pending` request in state. `useCalculator` sends that request and dispatches the result or the
error back. So the reducer can be tested with no mocks, and the hook only handles the network. Clearing
or unmounting aborts the request in flight (`AbortController`), so a slow response can never overwrite
a newer state.

**Errors are mapped once on each side.** The API returns stable codes. The TypeScript client turns each
code into a friendly message ("Can't divide by zero."), handles network failures and malformed
responses, and falls back to the server's message for codes it doesn't know. Components only render.

**Logging.** One zap logger, built in `main`, writes one line per request with `request_id`, method, path,
status, latency, client IP and user agent. Level follows status (2xx Info, 4xx Warn, 5xx Error).
Errors are logged once, at the edge: the service never logs, and handlers attach errors to the Gin
context for the logging middleware. Panics are recovered, logged with a stack trace, and returned as a
generic `INTERNAL_ERROR` that hides internal details.

**API docs from a hand-written OpenAPI 3.1 spec**, embedded in the binary with `go:embed` and served with
Swagger UI. I chose this over swaggo annotations because it needs no code generator or extra
dependencies, and produces OpenAPI 3.1 rather than Swagger 2.0. The downside of hand-written docs is
drift, which a router test catches by comparing the spec against the registered routes.

**Docker.**
- The backend is a static binary on `distroless/static`: no shell, runs as non-root, about 37 MB. Since
  there's no `curl`, the health check runs the binary itself (`/server -healthcheck`).
- docker compose runs the backend container read-only with no Linux capabilities.
- nginx serves the built frontend, sends unknown paths to the app, and proxies `/api`, `/healthz` and
  `/swagger`. The browser sees a single origin, so no CORS is needed in production. CORS stays
  configurable for setups that serve the frontend from another origin.

**Minimal dependencies.** Backend: gin, zap, testify. Frontend: react, plus vite, vitest, Testing Library
and oxlint for development. No UI framework: plain CSS with a grid keypad.

---

## Assumptions

- **Percentage is `a % of b`** (`{"a":15,"b":200}` → 30), the most common calculator meaning. It is a
  binary operation, not "divide by 100".
- The calculator evaluates one operation at a time, left to right (`2 + 3 × 4` = 20). Operator
  precedence and full expressions were out of scope.
- There is no history, persistence or authentication: the assignment ranks correctness and clarity above
  features.
- Inputs are capped at 15 digits so every typed number is stored exactly as a float64. Results can be
  larger and are shown in scientific notation (`1e+21`).
- `power(0, 0)` returns `1` (Go's `math.Pow` and IEEE-754 convention).
- The UI is English only, and numbers use `.` as the decimal separator (`,` is accepted from the keyboard).

---

## What I'd do with more time

- **CI pipeline** (GitHub Actions): vet, lint and tests with coverage for both apps, then build the images
  and smoke-test the stack. Today these checks run locally.
- **Browser end-to-end tests** (Playwright) against `docker compose`, including the mobile layout.
- **Generate the TypeScript API types from the OpenAPI spec**, so frontend and backend can't drift either.
- **A Content-Security-Policy header**: needs Swagger UI's assets to be self-hosted (they come from a
  pinned jsDelivr version today).
- **Production hardening**: rate limiting, Prometheus metrics and OpenTelemetry tracing (the request ID
  is already in place), and an unprivileged nginx image.
- **A decimal mode** (e.g. `shopspring/decimal`) for exact base-10 arithmetic, the right choice for money.
- **Expression input** with operator precedence and parentheses, parsed and evaluated on the server.
- Calculation history, locale-aware number formatting, and an installable PWA.

---

## AI usage

This project was built with Claude Code (Claude Opus) as a pair programmer. Every prompt I gave it,
with a short note on what it produced, is in [`PROMPTS.md`](PROMPTS.md).
