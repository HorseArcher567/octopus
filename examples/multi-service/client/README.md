# Multi-Service Client Example

## Run

```bash
go run . -config config.yaml \
  -target etcd:///multi-service-demo \
  -api http://127.0.0.1:8090/hello
```

## Flags

- `-config`: client config path
- `-target`: gRPC target (`etcd:///service-name` or `host:port`)
- `-api`: HTTP API endpoint for health check

## What It Does

1. Build gRPC client from config
2. Call `CreateUser`, `GetUser`, `CreateOrder`, `ListProducts`
3. Call HTTP `/hello` endpoint
