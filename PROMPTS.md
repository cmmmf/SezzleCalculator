# Prompts used

The assignment allows any AI tooling and asks to share the prompts. This project was built with
**Claude Code** (Claude Opus 5.5). Below are the prompts I gave it, in order, lightly trimmed for length.
Project conventions the agent followed are in the planning `CLAUDE.md` (summarized in the README's design section).

## 1. Planning

> lets create a plan to execute this assessment
>
> *(followed by the full assignment text from Sezzle)*

> i want to use model service controller arquitecture, i want both backend and frondend to be dockerized, use gin
> to handle REST communication and zap logger to logs

> what are your suggestions

Claude proposed: one REST route per operation, all three optional operations, `float64` with a
documented note on decimal money handling, nginx-proxied frontend container, and phase-by-phase commits.

> go ahead with your suggestions and write CLAUDE.md

## 2. Implementation

> I created the repository: https://github.com/cmmmf/SezzleCalculator

> lets go

Phase 2: Claude wrote the config, zap logger, API models and the calculator service with table-driven
tests. It proposed splitting NaN results (`UNDEFINED_RESULT`) from overflow (`RESULT_OUT_OF_RANGE`) and
normalizing `-0` to `0`.

> yes

Phase 3: Claude wrote the Gin controller (strict JSON binding: unknown fields and trailing data rejected),
middleware (request ID, zap request logger, recovery, CORS, body limit), router and `main.go` with graceful
shutdown and a `-healthcheck` flag for the distroless image. It smoke-tested the binary and fixed two log
issues it found (noisy stack traces on 4xx, Gin's error formatting).

> Add swagger to my back end service

Claude hand-wrote an OpenAPI 3.1 spec (`internal/docs/openapi.yaml`), embedded it with `go:embed` and served
it with Swagger UI at `/swagger`. It chose this over swaggo to avoid a code generator and new dependencies
(swaggo also only emits Swagger 2.0), and added a router test that fails if the spec and the registered
routes drift apart.

> how can i start the server?

> let's start phase 4

Phase 4: Claude wrote the typed API client (`calculatorService.ts`, injectable `fetch`, error codes mapped to
user-friendly messages, network failures and malformed responses handled), a pure keypad reducer (pressing
`=` sets a `pending` request instead of fetching; operator chaining, sign toggle, 15-digit input cap, client
validation), a display formatter and the `useCalculator` hook that sends pending requests and aborts them on
clear/unmount. Running the tests caught a real bug: formatting to 12 significant digits rounded
15-digit inputs, so it moved to 15 (the float64 round-trip limit).

> do it

Phase 5: Claude replaced the Vite scaffold with `Display`, `Keypad` (accessible names on symbol keys, `=`
disabled in flight), `ErrorBanner` and `Calculator`, a `useKeyboardShortcuts` hook (ignores Ctrl/Alt/Meta
shortcuts, prevents Enter from also clicking a focused key), and plain responsive CSS with a 4-column grid
and dark mode. User-flow tests drive the real client, hook and reducer against a fake `fetch` backend. It
checked the layout in headless Edge at desktop, 390 px and 320 px widths.
