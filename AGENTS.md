# AGENTS.md

## Overview

Security/master-data microservice for Stratifyr: securities, industries, metrics,
market days/holidays, security stats, market-data jobs. Served over HTTP REST + gRPC
using the GoFr framework, backed by MySQL and Redis.

For framework details (config loading, migrations, context datasources, etc.),
refer to https://gofr.dev/AGENTS.md

## Commands

- Run dev server: `go run .` — loads `configs/.env` (HTTP :8000, gRPC :9000)
- Verify changes: `go build ./... && go vet ./...` (no tests or lint config exist)

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
  (`internal/services/securityMetric.go`); stores invalidate related keys on
  Create/Update — preserve invalidation when touching store writes.

## Conventions

- Service errors: `&ErrResp{Code, Message}` (`internal/services/error.go`,
  implements `StatusCode()`); stores wrap DB errors in `datasource.ErrorDB`.
- Schema changes: add a `migration.Migrate` entry keyed by Unix timestamp to
  `migrations/000_all.go` (see `1742025361.go`); migrations auto-run on startup
  via `app.Migrate()`.
- Config files: `configs/.env` is the committed template, `configs/.test.env`
  points at local infra (MySQL :3306, Redis :6379), `*.local.env` is gitignored.
