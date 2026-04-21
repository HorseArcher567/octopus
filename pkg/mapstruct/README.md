# mapstruct

`pkg/mapstruct` provides a small decoder for mapping `map[string]any` values into Go structs.

## Features

- struct decoding from `map[string]any`
- field matching by tag or field name
- optional strict mode
- configurable time layout
- nested structs, pointers, slices, arrays, and maps

## Usage

```go
package main

import (
    "fmt"

    "github.com/HorseArcher567/octopus/pkg/mapstruct"
)

type User struct {
    Name string `yaml:"name"`
    Age  int    `yaml:"age"`
}

func main() {
    input := map[string]any{
        "name": "alice",
        "age":  18,
    }

    var user User
    err := mapstruct.New().Decode(input, &user)
    if err != nil {
        panic(err)
    }

    fmt.Println(user.Name, user.Age)
}
```

## Options

```go
mapstruct.New().
    WithTagName("json").
    WithStrictMode(true).
    WithTimeLayout("2006-01-02 15:04:05")
```

## Notes

- `New()` currently uses the `yaml` tag by default.
- If a field cannot be decoded, strict mode returns an error; otherwise the field is skipped.
- `ErrArrayLengthMismatch` is always returned when array input length does not match the target array length.
