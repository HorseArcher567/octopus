# Multi-Service Client Example

See the project overview in the repository root [`README.md`](../../../README.md).

This client example runs as a short-lived Octopus app whose primary runtime service is the in-process job scheduler.
Each registered job represents one self-contained RPC or HTTP demo scenario.
The scheduler runs all registered jobs concurrently, and the app exits naturally after they complete.

## Run

```bash
go run . -config config.yaml \
  -target etcd:///multi-service-demo \
  -api http://127.0.0.1:8090/hello
```

## Flags

- `-config`: client config path
- `-target`: gRPC target (`etcd:///service-name`, `direct:///host:port[,host:port]`, or `host:port`)
- `-api`: HTTP API base URL or `/hello` endpoint

## Structure

- `main.go`: process entrypoint
- `internal/jobs`: job registration and scenario implementations
- `config.yaml`: client infrastructure config

## Registered Jobs

- `rpc.user_flow`: create and then get a user over gRPC
- `rpc.order_flow`: create a user and then create an order over gRPC
- `rpc.product_flow`: list products over gRPC
- `http.hello`: call HTTP `/hello`
- `http.user_flow`: create and then get a user over HTTP
- `http.order_flow`: create a user and then create an order over HTTP
- `http.product_flow`: list products over HTTP

## What It Does

1. Build a short-lived Octopus app from config
2. Register RPC and HTTP demo scenarios as jobs
3. Run the app so the job scheduler executes those scenarios concurrently
4. Exit naturally after the jobs complete
