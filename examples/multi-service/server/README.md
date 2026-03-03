# Multi-Service Server (Best Practice Example)

## Run

```bash
# optional
cp .env.example .env
export $(grep -v '^#' .env | xargs)

go run ./cmd/server -config configs/config.dev.yaml
```

## Structure

- `cmd/server`: process entrypoint
- `internal/bootstrap`: module lifecycle (infra/rpc/api modules)
- `internal/domain`: domain models
- `internal/repository`: data access layer
- `internal/service`: business service layer
- `internal/transport/grpc`: gRPC handlers and mapping
- `internal/transport/http`: HTTP routes
- `configs`: development and template configs

## Lifecycle

The example uses 3 `app.Module`s:

1. `infra`: initialize DB and build repositories
2. `rpc`: build business services from infra repositories and register gRPC handlers
3. `api`: register HTTP routes

Startup sequence in `cmd/server/main.go`:

```go
infra := bootstrap.NewInfraModule()
app.MustRun(configFile, []app.Module{
    infra,
    bootstrap.NewRPCModule(infra),
    bootstrap.NewAPIModule(),
})
```

## E2E Test

```bash
OCTOPUS_TEST_MYSQL_DSN='root:123456@tcp(127.0.0.1:3306)/octopus?charset=utf8mb4&parseTime=True&loc=Local' \
GOCACHE=/tmp/go-build go test ./tests/e2e -v
```
