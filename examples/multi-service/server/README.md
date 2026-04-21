# Multi-Service Server (Reference Example)

See the project overview in the repository root [`README.md`](../../../README.md).

This example is a single-process reference application.
It demonstrates how one Octopus app exposes multiple business services over both gRPC and HTTP.

## Run

```bash
go run . -config config.yaml
```

If you need to initialize the schema manually:

```bash
mysql -uroot -p octopus < schema.sql
```

## Structure

- `main.go`: process entrypoint
- `internal/user`: user business module
- `internal/order`: order business module
- `internal/product`: product business module
- `internal/shared`: small shared assembly helpers
- `configs`: development and template configs

Each business module keeps its own assembly, repository, service, and HTTP/gRPC transport code together.
This keeps the example organized by business capability instead of global technical layers.

## Assembly

The example is assembled through business-capability actions:

- `shared.AssembleHello`
- `user.Assemble`
- `order.Assemble`
- `product.Assemble`

Application entry in `main.go`:

```go
a, err := assemble.Load(
    configFile,
    assemble.With(
        shared.AssembleHello,
        user.Assemble,
        order.Assemble,
        product.Assemble,
    ),
)
if err != nil {
    return err
}

return a.Run(ctx)
```

Each module assembles its own repository, service, and HTTP/gRPC registrations from shared infrastructure dependencies in `store.Store`.
Business objects are not written back into the store as container-managed dependencies.

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

The e2e test applies `schema.sql` automatically before startup.
