# pkg/resource

`pkg/resource` manages named infrastructure resources shared by modules.

Today it supports:

- MySQL connections
- Redis clients
- startup health checks
- ordered shutdown through `Close()`
