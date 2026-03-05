# go-binlog-sync

## Overview

go-binlog-sync is a binlog sync tool. It reads binlog from source mysql and syncs to destination mysql.

- cmd/gbs is the main entry point.
- pkg/logger is the logger package.
- pkg/context contains the configuration context, all configurations are loaded from this context.


## Technology Stack

- Go: 1.26

## Core Rules

- Follow Dave Cheney's Go style: clarity over cleverness; make code obvious.
- Keep functions small and focused; if a function needs a long comment to explain, simplify it.
- Prefer simple, readable control flow; avoid deep nesting; return early on errors.
- Handle errors explicitly; wrap with context using `fmt.Errorf("...: %w", err)` when returning.
- Avoid unnecessary abstraction; only introduce interfaces when there is a clear use case.
- Name things for their purpose; avoid stutters and redundant package qualifiers.
- Keep packages cohesive and minimal; avoid importing large packages for small helpers.
- Use zero values and `var` when they improve clarity; avoid `new(T)` unless required.
- Prefer `make` with explicit lengths/capacities when it clarifies intent.
- No unused parameters, variables, or imports.
- Avoid global state unless it is immutable or clearly owned (e.g., config/logging).
- Tests: table-driven where it improves clarity; include edge cases; keep test names descriptive.