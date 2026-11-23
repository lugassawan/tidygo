# ifacename

Requires a single-method interface to be named after its method plus an agent-noun suffix, the way the standard library does (`Read` → `Reader`, `Close` → `Closer`). The check is purely syntactic.

An interface is checked when its body is exactly one exported method. Interfaces that embed another interface, contain a type-set element (`~int | ~string`), declare more than one method, or have an unexported sole method (the sealed-interface pattern, such as `type Node interface{ isNode() }` or `type Option interface{ apply(*cfg) }`) are skipped. Local (function-scope) and generic interfaces are checked too.

Accepted spellings are compared ignoring the case of the first letter. For a method `w`:

| Rule | Example |
|---|---|
| `w` + `er` | `Read` → `Reader`, `String` → `Stringer`, `Marshal` → `Marshaler` |
| `w` ending in `e`: `w` + `r`, or drop the `e` and add `or` | `Close` → `Closer`, `Validate` → `Validator`, `Execute` → `Executor` |
| `w` + `or` | `Visit` → `Visitor`, `Edit` → `Editor` |
| consonant + `y`: replace `y` with `ier` | `Copy` → `Copier`, `Notify` → `Notifier` |
| consonant-vowel-consonant ending (not `w`, `x`, `y`): double the last letter | `Stop` → `Stopper`, `Run` → `Runner`, `Get` → `Getter` |

The same rules apply to the method's first camel-case word with the rest appended unchanged, which covers the standard library's `ReadFrom` → `ReaderFrom`, `WriteTo` → `WriterTo` and `ReadAt` → `ReaderAt`.

```go
// Bad
type FileRead interface{ Read() }      // reported, suggests Reader
type Handler interface{ ServeHTTP() }  // reported, suggests ServerHTTP (see allowed-names)
type closeable interface{ Close() }    // reported, suggests closer

// Good
type Reader interface{ Read() }
type Closer interface{ Close() error }
type ReaderFrom interface{ ReadFrom() }
type Getter[T any] interface{ Get() T }
```

The diagnostic's suggested name is a heuristic example, not the only accepted spelling. For multi-word methods it can read awkwardly (`RoundTrip` suggests `RounderTrip`), and stress-final verbs get an unusual spelling (`Visit` suggests `Visiter`, `Format` suggests `Formater`, `Commit` suggests `Commiter`). The suggestion is always itself an accepted name, and the standard spelling (`Visitor`, `Formatter`, `Committer`) is accepted too.

## Known limitations

The rule is deliberately strict, so a few common, idiomatic shapes are reported and need `allowed-names`:

- **Multi-word methods** are matched by suffixing the *first* word only (`GetUser` → `GetterUser`), so `UserGetter` is reported. The same applies to the "object + agent noun" names in the standard library: `ByteReader` (`ReadByte`), `RuneReader`, `StringWriter` (`WriteString`), `TextMarshaler`, and `Marshaler` for `MarshalJSON`.
- **Exported marker methods** are still reported, because only an *unexported* sole method is skipped: `type Expr interface{ Expr() }` suggests `Exprer` and `type EventMarker interface{ IsEvent() }` suggests `IserEvent`.
- **Type aliases** (`type X = interface{ Read() }`) are checked like any other declaration.
- **`Error() string`** interfaces, because `Error` → `Errorer`.

## Dependencies

Only interfaces declared in the packages being linted are inspected. Interfaces declared in dependencies or the standard library (`io.ReaderFrom`) are never reported, and neither is *using* them (embedding one, or taking one as a parameter), because that is a reference, not a declaration. The only exposure is your own code declaring a name that mirrors a standard library one. The forms that follow the rules above (`Reader`, `ReaderFrom`, `Stringer`) pass; the ones that do not (see Known limitations) need `allowed-names`.

## Options

| Option | Default | Description |
|---|---|---|
| `allowed-names` | `[]` | Interface names exempt from the check, for names that cannot be derived from the method (e.g. `Handler` for `ServeHTTP`, `ByteReader` for `ReadByte`). Matched exactly. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          ifacename:
            allowed-names:
              - Handler
              - ByteReader
```
