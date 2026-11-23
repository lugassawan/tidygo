# nolocalstruct

Forbids named struct type declarations inside function bodies. Structs should be declared at package level. Non-struct local types (e.g. `type myInt int`) are allowed unless `all-types` is enabled.

```go
// Bad
func process() {
    type Request struct{ URL string } // error: declare at package level
    _ = Request{}
}

// Good
type Request struct{ URL string }

func process() {
    _ = Request{}
}

// OK — non-struct local types are allowed
func convert() {
    type myInt int
    _ = myInt(0)
}
```

## Options

| Option | Default | Description |
|---|---|---|
| `all-types` | `false` | Also report local interfaces, defined types (`type id int`, `type fn func()`) and generic types. |

Off by default, so existing configurations see no new findings. Aliases (`type X = int`) are not new named types and stay allowed, except an alias of a struct (`type X = struct{}`), which is reported as it always was. With it enabled:

```go
// Bad
func convert() {
    type myInt int          // error: named type "myInt" should be declared at package level
    type Reader interface{} // error: named type "Reader" should be declared at package level
}
```

Struct declarations keep the original message (`named struct type %q ...`) in both modes, so existing `nolint` directives and exclusion rules keep matching.

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          nolocalstruct:
            all-types: true
```
