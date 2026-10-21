## 1. Planning
contract, error codes, logging and testing rules, and an 8-phase plan that the rest of this log follows.

> Let's create a plan to complete this assessment.
>
> *(followed by the full assignment text from Sezzle, summarized below)*

> I want to use a model–service–controller architecture, with both the backend and the frontend dockerized.
> Use Gin for the REST layer and zap for logging.
> Do it as simple as possible

### Requirements

This table summarizes the assignment text I pasted. The requirements we added during planning are
listed further down, where `CLAUDE.md` was written.

**From the assignment**

| | Requirement |
|---|---|
| **FR1** | Arithmetic operations: add, subtract, multiply, divide |
| **FR2** | Optional operations: power, square root, percentage (all three implemented) |
| **FR3** | Frontend in React (TypeScript) with an intuitive calculator UI |
| **FR4** | Frontend input validation and error handling |
| **FR5** | Basic mobile support |
| **FR6** | Backend in Go exposing REST endpoints for the operations |
| **FR7** | Backend input validation and edge cases: division by zero, invalid data |
| **FR8** | JSON responses |
| **NFR1** | Judged on correctness, clean and idiomatic code, and testability |
| **NFR2** | Unit tests for frontend and backend, with a coverage report |
| **NFR3** | README: setup, how to run each side, API call examples, design decisions and assumptions |
| **NFR4** | Optional: Docker for the full stack (implemented) |
| **NFR5** | Time budget of about 2–4 hours |
| **NFR6** | Deliverables: Git repository link, README, and the AI prompts used |

> What are your suggestions?

Claude proposed:
- one REST route per operation;
- implementing all three optional operations;
- `float64` numbers, with a README note on how money would be handled instead (decimal or integer cents);
- an nginx container serving the frontend and proxying the API;
- one commit per phase.

> Go ahead with your suggestions and create CLAUDE.md with the plan, including all the functional and
> non-functional requirements.

Claude wrote `CLAUDE.md`, which I reviewed and kept as the project plan: the functional and non-functional
requirements, folder layout, API contract and error codes, logging rules, testing targets and an 8-phase
checklist. Phase 1 came next: scaffolding the Go module and the Vite app, `.gitignore`, and this file.

The tables below list the requirements written into `CLAUDE.md` on top of the assignment's. They combine
my constraints (architecture, Gin, zap, Docker for both sides) with Claude's suggestions that I accepted.

**Agreed functional requirements**

| | Requirement |
|---|---|
| **AFR1** | One REST route per operation: `POST /api/v1/{add,subtract,multiply,divide,power,percentage,sqrt}` with body `{"a": …, "b": …}` (`sqrt` takes only `a`); percentage means **a% of b** |
| **AFR2** | Success response: `200 {"operation": "...", "result": ...}` |
| **AFR3** | Error response: `{"error": {"code": "...", "message": "..."}}`, with one stable code per situation |
| **AFR4** | Malformed JSON, missing field, wrong type, unknown field or non-finite operand → `400 INVALID_REQUEST`. A missing operand is never read as `0` |
| **AFR5** | Division by zero → `422 DIVISION_BY_ZERO`; square root of a negative → `422 NEGATIVE_SQUARE_ROOT` |
| **AFR6** | Infinite result (e.g. `power(10, 400)`) → `422 RESULT_OUT_OF_RANGE`; undefined result (NaN, e.g. `power(-8, 0.5)`) → `422 UNDEFINED_RESULT` |
| **AFR7** | Unknown route → `404 NOT_FOUND`; wrong method → `405 METHOD_NOT_ALLOWED`; body over 1 KB → `413 PAYLOAD_TOO_LARGE`; panic → `500 INTERNAL_ERROR` |
| **AFR8** | Numbers are `float64` with no rounding in the API; `-0` is returned as `0` |
| **AFR9** | `GET /healthz` returns `{"status": "ok"}` |
| **AFR10** | Calculator keypad flow: enter a, pick an operator, enter b, press `=`, then call the API; square root and percentage handled explicitly |
| **AFR11** | Keyboard input support |
| **AFR12** | Frontend validation before calling the API: one decimal point, a maximum input length, no empty operand |
| **AFR13** | Loading state while a request is in flight (`=` disabled), backend errors shown in an error banner, network failures handled |
| **AFR14** | API error codes mapped to user-friendly messages |
| **AFR15** | Display formatting done in the frontend |

**Agreed non-functional requirements**

| | Requirement |
|---|---|
| **ANFR1** | Model / service / controller architecture: router → controller → service, never the other way |
| **ANFR2** | The service is pure business logic (no Gin, HTTP or logging) behind an interface, returning typed errors; the controller depends on the interface so tests can inject a fake |
| **ANFR3** | Dependencies injected through constructors; no globals, including the logger; `main.go` only wires things together |
| **ANFR4** | Gin for REST, zap for structured logging |
| **ANFR5** | Logging: one injected logger; JSON in containers, console locally; per request: request ID, method, path, status, latency, client IP, user agent |
| **ANFR6** | Log levels by status: 2xx/3xx Info, 4xx Warn, 5xx Error. Each error logged once, at the edge; panics logged with the stack |
| **ANFR7** | Request ID taken from `X-Request-ID` or generated, and echoed back in the response |
| **ANFR8** | Request body limited to 1 KB; CORS origins configurable |
| **ANFR9** | Configuration through environment variables: `PORT`, `GIN_MODE`, `LOG_LEVEL`, `LOG_FORMAT`, `CORS_ORIGINS` |
| **ANFR10** | Graceful shutdown, flushing the logger on exit |
| **ANFR11** | Frontend: Vite + React + strict TypeScript, plain CSS, no UI framework; responsive CSS grid keypad for mobile |
| **ANFR12** | Frontend layering: a pure keypad state machine (no fetch), a hook for side effects, components that only render |
| **ANFR13** | Backend tests: testify, table-driven, `httptest`, a fake service for controller tests, log assertions for middleware; at least 90% coverage on `internal/` |
| **ANFR14** | Frontend tests: Vitest + React Testing Library + v8 coverage for the state machine, the API client (mocked `fetch`) and user flows |
| **ANFR15** | Every API edge case covered by a test; generated coverage output not committed |
| **ANFR16** | Docker: backend built to a minimal non-root distroless image; frontend built and served by nginx, which proxies the API |
| **ANFR17** | docker compose runs both images; the frontend starts only when the backend's health check passes |
| **ANFR18** | Priorities: correctness > clarity > maintainability > features. No auth, persistence or history; extra ideas go in the README instead |
| **ANFR19** | Minimal dependencies: gin, zap, testify (Go); react, vite, vitest, testing-library (frontend) |
| **ANFR20** | Code quality: `gofmt`/`go vet` clean, errors wrapped with `%w`, no ignored errors; no `any` or unused code in TypeScript |
| **ANFR21** | One commit per phase with descriptive messages, pushed to GitHub |

## 2. Implementation

> I've created the repository: https://github.com/cmmmf/SezzleCalculator

> Let's start.

### Phase 2 — core backend

Claude wrote:
- the config (environment variables);
- the zap logger;
- the API models;
- the calculator service, with table-driven tests.

It proposed reporting undefined results (NaN) separately from too-large ones (±Inf), and returning `-0`
as `0`.

> Yes, go ahead.

### Phase 3 — HTTP layer

Claude wrote:
- the Gin controller, with strict JSON parsing that rejects unknown fields and trailing data;
- middleware for request IDs, zap request logging, panic recovery, CORS and a 1 KB body limit;
- the router;
- `main.go`, with graceful shutdown and a `-healthcheck` flag for the minimal Docker image.

It ran the binary and fixed two logging problems it found: stack traces were being logged on 4xx
responses, and Gin's error text was badly formatted.

### Swagger

> Add Swagger documentation to the backend service.

Claude:
- wrote an OpenAPI 3.1 spec by hand (`backend/internal/docs/openapi.yaml`);
- built it into the binary with `go:embed` and serves it with Swagger UI at `/swagger`;
- added a router test that fails if the spec and the registered routes drift apart.

It chose this over swaggo because swaggo needs a code generator and extra dependencies, and only produces
Swagger 2.0.

### Phase 4 — frontend logic

> Let's start phase 4.

Claude wrote:
- a typed API client that turns error codes into user-friendly messages and handles network failures;
- a pure keypad state machine: pressing `=` records a pending request instead of fetching;
- a display formatter;
- the `useCalculator` hook, which sends pending requests and cancels them on clear.

Running the tests caught a real bug: rounding the display to 12 significant digits changed 15-digit
inputs, so the formatter now uses 15 digits (the most a float64 always stores exactly).

### Phase 5 — UI

> Go ahead with phase 5.

Claude:
- replaced the Vite scaffold with the `Display`, `Keypad`, `ErrorBanner` and `Calculator` components;
- added keyboard support;
- wrote a responsive CSS grid layout with dark mode;
- wrote user-flow tests that run the real client and state logic against a fake backend.

It checked the layout in a headless browser at desktop, 390 px and 320 px widths.

### Phase 6 — Docker

> Commit, then start phase 6.

Claude added:
- the frontend Dockerfile (Node build, then nginx);
- `nginx.conf`: app fallback; `/api`, `/healthz` and `/swagger` proxied to the backend; cache and
  security headers;
- `docker-compose.yml`: the backend's health check runs the server binary itself, and the frontend starts
  only once the backend is healthy.

It ran the stack locally and checked each route through nginx.

> Set the backend port to 5000 and the frontend port to 8080.

Claude changed the docker compose host ports to 5000 (backend) and 8080 (frontend).

> You don't need to create GitHub workflows.

Claude removed the GitHub Actions workflow it had added in phase 6. Tests, lint and coverage run locally
(commands in the README), and a CI pipeline is listed under "What I'd do with more time".

### Phase 7 — README

> Let's move on to phase 7.

Claude wrote the README:
- setup, and running the project with Docker or locally;
- API reference, error codes, and curl/PowerShell examples;
- calculator usage and keyboard shortcuts;
- test and coverage commands;
- design decisions, assumptions, and "what I'd do with more time".

It ran every curl example against a fresh build of the server to confirm the documented outputs.

> Commit CLAUDE.md as well.

## 3. Final review

### Phase 8 — review against the deliverables

> Check the repo against the assignment's deliverables and review the code as a strict reviewer; list
> issues before fixing.

Claude:
- checked each requirement and deliverable from the assignment against the repo;
- read the backend, frontend, Dockerfiles, nginx and compose configuration;
- ran `go vet`, the Go tests, the frontend typecheck, lint and tests with coverage. Everything passed:
  backend `internal/` packages at 94–100% statement coverage, frontend 119 tests with 100% line
  coverage.

It listed these issues without fixing them:
1. `PROMPTS.md` is still gitignored, so the README's "AI usage" link points to a file that isn't in the
   repository. One of the three deliverables is missing until it is committed.
2. When the backend is down, nginx returns an HTML 502, and the frontend shows "unexpected response"
   instead of "can't reach the calculator service".
3. Under docker compose, `client_ip` in the backend logs is always the nginx container, because the
   backend trusts no proxies and ignores the `X-Forwarded-For` header nginx sets.
4. `CLAUDE.md` doesn't mention the Swagger endpoint or the `docs` package, which were added after the
   plan was written.
5. The end-to-end `docker compose up` check and the push from phase 8 are still to do.
