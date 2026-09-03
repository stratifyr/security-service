# AGENTS.md

## Secrets guard (READ FIRST)

- `configs/*.local.env` contains deploy secrets (DB/Redis creds) and is
  gated by `.opencode/opencode.json` (`read: **/*.local.env: deny`).
  NEVER read or print it — including via shell commands (`cat`, `sed`,
  `head`, copy, redirect). Use the file `read` tool only on allowed files.
  `*.local.env` is gitignored; only committed templates `.env` and
  `.test.env` are safe to read.
- Never commit secrets, tokens, or a `.local.env` copy.

## Overview

Security/master-data microservice for Stratifyr: securities, industries, metrics,
market days/holidays, security stats, market-data jobs. Served over HTTP REST + gRPC
using the GoFr framework, backed by MySQL and Redis.

For framework details (config loading, migrations, context datasources, etc.),
refer to https://gofr.dev/AGENTS.md

## Commands

- Run dev server: `make run` — loads `configs/.env` (HTTP :8000, gRPC :9000)
- Verify changes: `make verify`

## Architecture

- Layers: `internal/handlers` → `internal/services` → `internal/stores`, wired
  manually in `main.go` (no DI). Add new endpoints there too.
- Stores are stateless structs; all MySQL/Redis access goes through
  `ctx.SQL` / `ctx.Redis` on `*gofr.Context`.
- gRPC: request/response types and the GoFr wrapper come from the external
  `github.com/stratifyr/security-service-proto` module — update protobufs in that
  repo, never here. gRPC methods delegate to `*GRPC` handlers via
  `internal/handlers/grpc.go`.
- Computed metric values are cached in Redis as msgpack
  (`internal/stores/securityMetric.go`); `securityStore` and `securityStatStore`
  invalidate related keys on Create/Update — preserve invalidation when touching
  those store writes.

## Conventions

- Service errors: `&Error{httpCode, message}` (`internal/services/error.go`,
  implements `StatusCode()`); stores wrap DB errors in `datasource.ErrorDB`.
- Schema changes: add a `migration.Migrate` entry keyed by Unix timestamp to
  `migrations/000_all.go` (see `1742025361.go`); migrations auto-run on startup
  via `app.Migrate()`.
- Config files: `configs/.env` is the committed template, `configs/.test.env`
  points at local infra (MySQL :3306, Redis :6379), `*.local.env` is gitignored.
