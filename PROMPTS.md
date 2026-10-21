# Prompts used

The assignment allows any AI tooling and asks to share the prompts. This project was built with
**Claude Code** (Claude Opus 5.5). Below are the prompts I gave it, in order, lightly trimmed for length.
Project conventions the agent followed are in the planning `CLAUDE.md` (summarized in the README's design section).

## 1. Planning

> lets create a plan to execute this assessment
>
> *(followed by the full assignment text from Sezzle)*

> i want to use service controller arquitecture, i want both backend and frondend to be dockerized, use gin
> to handle REST communication and zap logger to logs

> what are your suggestions

Claude proposed: one REST route per operation, all three optional operations, `float64` with a
documented note on decimal money handling, nginx-proxied frontend container, and phase-by-phase commits.

> go ahead with your suggestions and write CLAUDE.md

## 2. Implementation

*(appended as the work progresses)*
