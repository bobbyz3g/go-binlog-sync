# go-binlog-sync

## Project

MySQL/MariaDB binlog sync tool.

- `cmd/gbs`: main entry
- `pkg/context`: configuration loading and access
- `pkg/logger`: logging

## Stack

- Go 1.26

## Rules

- Follow Dave Cheney's Go style: prefer clarity over cleverness.
- Keep functions short, control flow simple, and return early on errors.
- Wrap returned errors with context using `fmt.Errorf("...: %w", err)`.
- Avoid unnecessary abstractions, especially interfaces without a clear use case.
- Use clear names; keep packages cohesive; remove unused params, vars, and imports.
- Prefer zero values and explicit `make(...)` sizes when they improve readability.
- Avoid mutable global state unless ownership is obvious, such as config or logging.
- Tests should be descriptive and cover edge cases; use table-driven tests when it helps.
