# CLAUDE.md — Sezzle take-home: full-stack calculator

This folder is **not** a study guide. The parent `InterviewsPrep/CLAUDE.md` rules (study-guide skeleton,
drills, self-check, etc.) **do not apply here**. Everything below overrides them for this folder.

## 0. The assignment (summary)

Sezzle asked for a full-stack calculator: **React (TypeScript) frontend + Go REST backend**, judged on
correctness, clean/idiomatic code, testability, unit tests + coverage, and documentation.
Budget: **~2–4 hours**. Deliverable: a Git repo link + README + the AI prompts used.

Required:
- Operations: add, subtract, multiply, divide. Optional (we do them): power, sqrt, percentage.
- Frontend: intuitive UI, input validation + error handling, basic mobile support.
- Backend: REST endpoints, input validation, edge cases (division by zero, invalid data), JSON responses.
- Unit tests for both layers + coverage report.
- README: setup, how to run frontend and backend, API call examples, design decisions/assumptions.
- Optional (we do it): Docker for the full stack.

**Priorities: correctness > clarity > maintainability > features.** No auth, persistence, history, or
other extras. Put ideas in the README's "What I'd do with more time" section instead.

## 1. Location and repo

- Project root: `Sezzle/sezzle-calculator/` (its own git repo). Don't put prep notes in it.
- Remote: https://github.com/cmmmf/SezzleCalculator (public, `origin`, branch `main`). Push after each phase.
- Go module: `github.com/cmmmf/SezzleCalculator/backend`.
- Commit after each phase (§8) with small, descriptive messages. Reviewers read the history.

```
sezzle-calculator/
├── backend/
│   ├── cmd/server/main.go       config → zap logger → wire layers → Gin → graceful shutdown
│   ├── internal/
│   │   ├── config/              env: PORT (8080), GIN_MODE, LOG_LEVEL, LOG_FORMAT (json|console), CORS_ORIGINS
│   │   ├── logger/              zap constructor from config
│   │   ├── model/               request/response DTOs, error response
│   │   ├── service/             CalculatorService interface + impl, domain errors
│   │   ├── controller/          CalculatorController (Gin handlers), error → HTTP mapping
│   │   ├── middleware/          request ID, zap request logger, zap recovery, CORS
│   │   └── router/              gin.Engine setup + route registration
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── services/calculatorService.ts   typed API client
│   │   ├── state/calculatorReducer.ts      pure keypad state machine
│   │   ├── hooks/useCalculator.ts          reducer + service call
│   │   └── components/                     Display, Keypad, ErrorBanner, …
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
├── README.md
└── PROMPTS.md
```

## 2. Architecture: service / controller

- Dependency direction: **router → controller → service**. Never the other way.
- **service**: pure business logic. No Gin, no `net/http`, no logging. Exposes an interface
  (`CalculatorService`) and returns typed errors (sentinel `var Err... = errors.New(...)`).
- **controller**: binds and validates the request (Gin binding), calls the service, maps errors to
  status + error code, writes JSON. Depends on the service **interface** so tests can inject a fake.
  Use one shared helper for the binary operations to avoid repeating bind/call/respond in every handler.
- **middleware**: cross-cutting concerns only.
- **main.go**: wiring only. Dependencies are injected through constructors; **no globals** (including the
  logger).

## 3. API contract

One route per operation:

| Method | Path | Body | Result |
|---|---|---|---|
| POST | `/api/v1/add` | `{"a": 10, "b": 4}` | 14 |
| POST | `/api/v1/subtract` | `{"a": 10, "b": 4}` | 6 |
| POST | `/api/v1/multiply` | `{"a": 10, "b": 4}` | 40 |
| POST | `/api/v1/divide` | `{"a": 10, "b": 4}` | 2.5 |
| POST | `/api/v1/power` | `{"a": 2, "b": 10}` | 1024 |
| POST | `/api/v1/percentage` | `{"a": 15, "b": 200}` | 30 (**a% of b**) |
| POST | `/api/v1/sqrt` | `{"a": 16}` | 4 |
| GET | `/healthz` | – | `{"status":"ok"}` |

- Success: `200 {"operation":"divide","result":2.5}`
- Error: `{"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}}`

| Situation | Status | Code |
|---|---|---|
| Malformed JSON, missing field, wrong type, non-finite operand | 400 | `INVALID_REQUEST` |
| Divide by zero | 422 | `DIVISION_BY_ZERO` |
| sqrt of negative | 422 | `NEGATIVE_SQUARE_ROOT` |
| Result is ±Inf (e.g. `power(10, 400)`, `power(0, -1)`) | 422 | `RESULT_OUT_OF_RANGE` |
| Result is NaN (e.g. `power(-8, 0.5)`) | 422 | `UNDEFINED_RESULT` |
| Unknown route | 404 | `NOT_FOUND` |
| Wrong HTTP method on a known route | 405 | `METHOD_NOT_ALLOWED` |
| Body over 1 KB | 413 | `PAYLOAD_TOO_LARGE` |
| Panic / unexpected | 500 | `INTERNAL_ERROR` |

- Operands are `*float64` with `binding:"required"` so a missing value is not mistaken for `0`.
- Limit the request body size (e.g. 1 KB).
- The service normalizes `-0` to `0`, so clients never see `-0`.
- Numbers are `float64`. **No rounding in the API.** The frontend handles display formatting. The README
  explains that money would use decimal or integer minor units instead.

## 4. Logging (zap)

- Build one `*zap.Logger` in `main.go` from config and inject it. JSON in containers, console locally.
- Request middleware logs: `request_id`, `method`, `path`, `status`, `latency`, `client_ip`, `user_agent`.
  Level: 2xx/3xx → Info, 4xx → Warn, 5xx → Error.
- The request ID comes from the `X-Request-ID` header or is generated, and is echoed back in the response.
- Recovery middleware logs panics with the stack and returns `INTERNAL_ERROR`.
- Log each error **once**, at the edge. The service never logs. Call `logger.Sync()` on shutdown.

## 5. Frontend conventions

- Vite + React + TypeScript (strict). No UI framework; plain CSS with a CSS grid keypad that works on mobile.
- UX: calculator keypad. Enter a, pick an operator, enter b, press `=`, then call the API. Unary sqrt and
  percentage are handled explicitly. Support keyboard input. Show a loading state and the backend error
  message in an error banner. Disable `=` while a request is in flight.
- Validation before calling the API: one decimal point, a max input length, no empty operand.
- `calculatorReducer` is pure (no fetch). `useCalculator` does the side effects. Components only render.
- The service maps API error codes to user-friendly messages and handles network failures.
- Dev: Vite proxies `/api` and `/healthz` to `http://localhost:8080`. Docker: nginx proxies to `backend:8080`.

## 6. Testing

- Go: `testify`, table-driven tests, `gin.SetMode(gin.TestMode)` + `httptest`, a fake service for
  controller tests, `zaptest/observer` for middleware log assertions. Target ≥ 90% coverage on `internal/`.
  - `go test ./... -race -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html`
- Frontend: Vitest + React Testing Library + `@vitest/coverage-v8`. Test the reducer, the service (mocked
  `fetch`) and user flows (`2 + 3 =` shows 5, `5 ÷ 0 =` shows the error).
- Every edge case in §3 must have a test.
- Don't commit generated coverage output. The README explains how to generate it.

## 7. Docker

- `backend/Dockerfile`: `golang:1.26-alpine` build (`CGO_ENABLED=0`) → `gcr.io/distroless/static`, non-root.
- `frontend/Dockerfile`: `node:24-alpine` build → `nginx:alpine` with `nginx.conf` (SPA fallback + `/api`
  proxy).
- `docker-compose.yml`: backend (`5000:8080`, healthcheck on `/healthz`) and frontend (`8080:80`,
  `depends_on: service_healthy`). `docker compose up --build` → http://localhost:8080.
- Because distroless has no shell or curl, the healthcheck uses the server binary's own `-healthcheck` flag.
- No CI workflow (the user doesn't want GitHub Actions). Tests, lint and coverage run locally.

## 8. Phases (check off as done)

- [x] 1. Scaffold, `git init`, `.gitignore`, `PROMPTS.md`
- [x] 2. config + logger + model + service, with tests
- [x] 3. controller + middleware + router + main, with tests
- [x] 4. Frontend: service, reducer, hook, with tests
- [x] 5. Components, responsive CSS, component tests
- [x] 6. Dockerfiles, nginx.conf, compose
- [x] 7. README: setup, running, curl examples, design decisions, assumptions, "more time" list
- [ ] 8. `docker compose up` end to end, final check, push, share the link

## 9. Working rules

- **Append each meaningful user prompt to `PROMPTS.md`** as we go (the assignment asks for it).
- After every phase: run the tests, commit, tick the box above.
- Keep dependencies minimal: gin, zap, testify (Go); react, vite, vitest, testing-library (frontend).
- Go: `gofmt`/`go vet` clean, wrapped errors with `%w`, `context` from the request, no ignored errors.
- TS: no `any`, no unused code, named exports.
- Before the push, re-check this file against the assignment's deliverables list.
