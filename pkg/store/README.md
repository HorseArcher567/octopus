# pkg/store

`pkg/store` provides a lightweight shared object store.

Features:

- store values by `name + exact type`
- typed retrieval via `Get[T](...)` and `GetNamed[T](...)`
- optional cleanup via `WithClose(...)`

Typical usage:

```go
s := store.New()

if err := s.SetNamed("primary", db, store.WithClose(db.Close)); err != nil {
    return err
}

repo := store.MustGet[*repository.UserRepository](s)
db := store.MustGetNamed[*database.DB](s, "primary")
```
