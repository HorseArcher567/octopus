# Multi-Service Server (Reference Example)

See the project overview in the repository root [`README.md`](../../../README.md).

This example is a single-process reference application.
It demonstrates how one Octopus app exposes multiple business services over both gRPC and HTTP.

## Run

Start the server with the example config:

```bash
go run . -config config.yaml
```

The demo now uses SQLite as its default business database:

- config section: `sqlite`
- resource name used by business domains: `primary`
- default DSN: `file:/tmp/octopus.db`

The example automatically applies `schema.sql` during startup through `assemble.WithStartupHooks(shared.InitSchema)`.
That means you do **not** need to initialize the schema manually before running the app.

If you want a clean local reset, simply remove the SQLite file and start again:

```bash
rm -f /tmp/octopus.db
go run . -config config.yaml
```

If you prefer an in-memory SQLite database for quick experiments, override the DSN:

```bash
SQLITE_DSN='file:octopus?mode=memory&cache=shared' go run . -config config.yaml
```

When running with in-memory SQLite, the embedded schema is still applied automatically during startup.

## Graceful Shutdown

The example `main.go` listens for `SIGINT` and `SIGTERM` and passes a cancelable context into `app.Run(...)`.
So pressing `Ctrl+C` triggers normal application shutdown instead of abruptly terminating the process.

## Structure

- `main.go`: process entrypoint
- `schema.sql`: SQLite-friendly example schema
- `internal/user`: user business domain
- `internal/order`: order business domain
- `internal/product`: product business domain
- `internal/shared`: shared setup and registration helpers
- `config.yaml`: example application config

Each business domain keeps its own registration, repository, service, and HTTP/gRPC transport code together.
This keeps the example organized by business domain instead of global technical layers.

## Domain registration

The example is assembled through business domains:

- app-level startup hook: `shared.InitSchema` via `assemble.WithStartupHooks(...)`
- `user.Register`
- `order.Register`
- `product.Register`

Application entry in `main.go`:

```go
a, err := assemble.Load(
    configFile,
    assemble.WithStartupHooks(shared.InitSchema),
    assemble.WithDomains(
        user.Register,
        order.Register,
        product.Register,
    ),
)
if err != nil {
    return err
}

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

return a.Run(ctx)
```

`shared.InitSchema` is an app-level startup hook that reads the primary database from the embedded `store.Reader` on `hook.Context` and applies the embedded `schema.sql`.
This makes both file-based and in-memory SQLite easy to use in the example.

Each business domain registers its own repository, service, and HTTP/gRPC endpoints from shared infrastructure dependencies resolved through the embedded `store.Reader` on `assemble.DomainContext`.
Business objects are not written back into the store as container-managed dependencies.

## Database config notes

`config.yaml` keeps both `sqlite` and `mysql` sections:

- `sqlite.primary` is the default business database used by the example
- `mysql.mysql-primary` is kept as a parallel DSN configuration example

This lets the config file act as both a runnable demo and a best-practice reference.

## Exposed Endpoints

The server exposes the same business domain over both transports:

- gRPC: `User`, `Order`, `Product` services
- HTTP:
  - `GET /users/:id`
  - `POST /users`
  - `GET /orders/:id`
  - `POST /orders`
  - `GET /products/:id`
  - `GET /products?page=1&page_size=10`
