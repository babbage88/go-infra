# go-infra Agent Guide

Keep this file short on purpose. It is for safe, efficient work in this package, not full architecture documentation.

## What This Service Is

`go-infra` is the Go backend for the infra dashboard and related automation. It exposes HTTP APIs, auth, RBAC, SSH/websocket features, and database-backed CRUD services.

## Stack

- Go 1.24
- PostgreSQL + `pgx`
- `sqlc` for generated queries in `database/infra_db_pg`
- Swagger/OpenAPI generation
- JWT auth
- Gorilla WebSocket

## Important Paths

- `main.go`, `startup.go`, `main_init.go`: startup and wiring
- `api/`: HTTP server and auth middleware/handlers
- `services/`: business/domain services
- `database/infra_db/`: DB pool setup
- `database/infra_db_pg/`: generated sqlc code
- `query.sql`: sqlc query source
- `migrations/`: schema changes
- `spec/`, `swagger.json`, `swagger.yaml`: API specs used by the UI client

## Working Rules

- Treat `database/infra_db_pg/**` as generated code. Edit `query.sql` or `migrations/` and regenerate instead.
- Keep handlers thin. Put business logic in `services/**` or focused helper packages.
- Match existing package organization instead of introducing a new layout style.
- Prefer adding to existing service areas such as `user_crud_svc`, `roles_service`, `host_servers`, or `user_secrets` when the change belongs there.
- Be careful with auth and permission changes: backend permission names are part of the UI contract.
- Avoid speculative cleanup in this repo. Many files are wired into Swagger generation, sqlc, or startup code.

## Common Workflows

Run locally:

```bash
go run . --local-development
```

Run tests:

```bash
go test ./...
```

Generate sqlc code:

```bash
sqlc generate
```

Create migration:

```bash
make new-sqlmigration
```

Apply migrations:

```bash
make apply-migration
```

Refresh Swagger artifacts when endpoint shapes change:

```bash
make local-swagger
```

## Integration Notes

- `db-helper-ui` depends on `swagger.json` to regenerate its client.
- If request/response structs or routes change, update Swagger artifacts and expect downstream UI compile fixes.
- SQL schema, generated queries, handlers, and Swagger must stay in sync.

## Editing Guidance

- For RBAC work, start in `services/roles_service/` and `api/user_api_handlers/`.
- For user CRUD, start in `services/user_crud_svc/`.
- For DB-backed behavior changes, inspect `query.sql` and the corresponding service together.
- For websocket/SSH work, inspect `services/ssh_connections/` before changing handlers elsewhere.
- For auth changes, trace through `api/authapi/` middleware and token helpers before editing.

## Verification

For most backend changes, run the smallest useful set:

```bash
go test ./...
```

If SQL changed, also run:

```bash
sqlc generate
```

If API contracts changed, also run:

```bash
make local-swagger
```

## More Context

Use `README.md` for longer project background. Keep this file focused on agent execution, file ownership, and regeneration workflows.
