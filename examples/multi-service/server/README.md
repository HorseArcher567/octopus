# Multi-Service Server (Reference Example)

See the project overview in the repository root [`README.md`](../../../README.md).

This example is a modular single process app, not a true multi-process microservice system.
It demonstrates how one Octopus app exposes multiple business services over both gRPC and HTTP.

## Run

```bash
# optional
cp .env.example .env
export $(grep -v '^#' .env | xargs)

go run ./cmd/server -config configs/config.dev.yaml
```

If you need to bootstrap the schema manually:

```bash
mysql -uroot -p octopus < schema.sql
```

## Structure

- `cmd/server`: process entrypoint
- `internal/bootstrap`: module lifecycle (infra/service/rpc/api modules)
- `internal/domain`: domain models
- `internal/repository`: data access layer
- `internal/service`: business service layer
- `internal/transport/grpc`: gRPC handlers and mapping
- `internal/transport/http`: HTTP routes
- `configs`: development and template configs

## Lifecycle

The example uses 4 `app.Module`s:

1. `infra`: initialize DB and build repositories
2. `service`: build business services from repositories
3. `rpc`: resolve services and register gRPC handlers
4. `api`: resolve services and register HTTP routes

Startup sequence in `cmd/server/main.go`:

```go
app.MustRun(configFile, []app.Module{
    bootstrap.NewInfraModule(),
    bootstrap.NewServiceModule(),
    bootstrap.NewRPCModule(),
    bootstrap.NewAPIModule(),
})
```

## Exposed Endpoints

The server exposes the same business capability over both transports:

- gRPC: `User`, `Order`, `Product` services
- HTTP:
  - `GET /hello`
  - `GET /users/:id`
  - `POST /users`
  - `GET /orders/:id`
  - `POST /orders`
  - `GET /products/:id`
  - `GET /products?page=1&page_size=10`

## E2E Test

```bash
OCTOPUS_TEST_MYSQL_DSN='root:123456@tcp(127.0.0.1:3306)/octopus?charset=utf8mb4&parseTime=True&loc=Local' \
GOCACHE=/tmp/go-build go test ./tests/e2e -v
```

The e2e test now applies `schema.sql` automatically before startup.
