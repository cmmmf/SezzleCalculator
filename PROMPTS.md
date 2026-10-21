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
