# pkg/resource

`pkg/resource` manages named shared infrastructure resources.

The runtime model is generic:

- `Register(kind, name, value, closeFn)`
- `Get(kind, name)`
- `MustGet(kind, name)`
- `List(kind)`
- `Close()`

Built-in resource kinds currently include:

- `resource.KindMySQL`
- `resource.KindRedis`

Helpers:

- `resource.As[T](...)`
- `resource.MustAs[T](...)`

This lets modules depend on resource capabilities without forcing framework core APIs to expose MySQL/Redis-specific methods.
