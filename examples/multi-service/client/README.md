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
- `-api`: HTTP API base URL or `/hello` endpoint

## What It Does

1. Build an Octopus app runtime from config
2. Call gRPC `CreateUser`, `GetUser`, `CreateOrder`, `ListProducts`
3. Call HTTP `/hello`
4. Call HTTP `POST /users`, `GET /users/:id`, `POST /orders`, `GET /products`
