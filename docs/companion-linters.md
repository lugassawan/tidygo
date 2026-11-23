# Companion linters

tidygo implements only conventions nothing else checks. Most of SonarLint's Go rules are already covered by linters that ship with golangci-lint. This page maps them, so you can turn on the right companions and so contributors can see what exists before proposing an analyzer.

**Contributors:** check this page before proposing a new analyzer. If a rule is listed below, enable the linter instead.

Checked against golangci-lint v2.15.0, the version in the README's `.custom-gcl.yml`, which bundles gosec v2.29.0, revive v1.17.1, go-critic v0.15.0 and staticcheck v0.8.2. Rule IDs shift between releases, so recheck them when you bump the version.

## Already covered by tidygo

| SonarLint rule | tidygo analyzer | Also available in |
|---|---|---|
| Functions should not have too many parameters | [`maxparams`](analyzers/maxparams.md) | revive `argument-limit` |
| Function names should comply with a naming convention | [`funcname`](analyzers/funcname.md) | revive `var-naming` (exempts `Test*`) |
| Functions should not have identical implementations | [`dupfunc`](analyzers/dupfunc.md) | `dupl` (token threshold) |

## Standard default set

These run in golangci-lint v2 without any config: `errcheck`, `govet`, `ineffassign`, `staticcheck` and `unused`. staticcheck runs all its checks except `ST1000`, `ST1003`, `ST1016`, `ST1020`, `ST1021` and `ST1022`.

| SonarLint rule | Linter |
|---|---|
| Variables should be used | `unused`, `ineffassign` |
| All code should be reachable | `govet` (unreachable) |
| Test functions should not call "t.Fatal" from separate goroutines | `govet` (testinggoroutine), `staticcheck` SA2002 |
| Deprecated "InterfaceData" method should not be used | `staticcheck` SA1019 |
| Identical expressions should not be used on both sides of a binary operator | `staticcheck` SA4000 |
| Variables should not be self-assigned | `govet` (assign), `staticcheck` SA4018 |
| Use "bytes.Equal" instead of "bytes.Compare" | `staticcheck` S1004 |
| Boolean literals should not be redundant | `staticcheck` S1002 (also revive `bool-literal-in-expr`) |
| Package imports should be consistent and avoid redundancy | `staticcheck` ST1019 (also revive `duplicated-imports`) |
| Nested blocks of code should not be left empty | `staticcheck` SA9003 (also revive `empty-block`) |

## Formatter

golangci-lint v2 runs these through `formatters`, not `linters`.

| SonarLint rule | Tool |
|---|---|
| Semicolons should not be used unnecessarily | `gofmt` |
| Empty statements should be removed | `gofmt` |
| Statements should be on separate lines | `gofmt` |

## Opt-in linters

Enable each one under `linters.enable`. The gocritic checkers listed here are also off by default, so add them to its `enabled-checks`.

| SonarLint rule | Linter |
|---|---|
| Context should not be stored in struct fields | `containedctx` |
| Context parameters should be reused instead of creating new background contexts | `contextcheck` |
| HTTP response bodies should be closed | `bodyclose` |
| Consecutive function parameters with the same type should be grouped | `gocritic` paramTypeCombine |
| Blank imports should be documented | revive `blank-imports` |
| Import statements should be factored into a single block | `grouper` (`import-require-single-import`) |
| Cognitive Complexity of functions should not be too high | `gocognit`, revive `cognitive-complexity` |
| Boolean checks should not be inverted | `gocritic` boolExprSimplify |
| Two / all branches should not have the same implementation | `gocritic` dupBranchBody, revive `identical-branches`, `identical-switch-branches` |
| Related "if/else if" statements should not have the same condition | revive `identical-ifelseif-conditions` |
| "switch" statements should have "default" clauses | revive `enforce-switch-style` |
| Control flow statements should not be nested too deeply | `nestif`, revive `max-control-nesting` |
| Functions and methods should not have too many lines | `funlen`, revive `function-length` |
| Files should not have too many lines of code | revive `file-length-limit` |
| Lines should not be too long | `lll`, revive `line-length-limit` |
| String literals should not be duplicated | `goconst`, revive `add-constant` |
| Unused function parameters should be removed | `unparam`, revive `unused-parameter` |
| Octal values should not be used | `gocritic` octalLiteral |
| Track uses of "TODO" / "FIXME" tags | `godox` |
| Track lack of copyright and license headers | `goheader`, revive `file-header` |
| Local variable and function parameter names should comply with a naming convention | revive `var-naming` (partial) |

`gocritic` dupBranchBody is on by default once you enable `gocritic`. paramTypeCombine, boolExprSimplify and octalLiteral are not.

## Security (`gosec`, opt-in)

Enable `gosec` under `linters.enable`. The G7xx rules are taint analyses and G124 is recent. All of them are in the gosec bundled with golangci-lint v2.15.0.

| SonarLint rule | gosec rule |
|---|---|
| Credentials / secrets should not be hard-coded | G101 |
| Constructing system command arguments from user input; OS command injection | G204, G702 |
| Zip slip | G305 |
| Cipher algorithms should be robust; weak hashing | G401, G405, G406, G501–G503 |
| Server hostname / certificate verification; weak TLS protocols | G402 |
| Cryptographic keys should be robust | G403 (RSA key length only, partial) |
| Pseudorandom number generators | G404 |
| CBC IVs should be unpredictable | G407 (hard-coded IV or nonce only, partial) |
| Insecure temporary files; publicly writable directories | G303 (predictable temp file path only, partial) |
| Loose POSIX file permissions | G301, G302, G306 |
| Debug features in production | G108 (pprof) |
| Cookies without HttpOnly / Secure | G124 |
| SQL injection; formatting SQL queries | G201, G202, G701 |
| Path injection | G304, G703 |
| SSRF | G107, G704 |
| Log injection | G706 |
| Open redirect | G710 |
| Clear-text protocols | revive `unsecure-url-scheme` (partial) |

tidygo's own security analyzers ([`aeadcipher`](analyzers/aeadcipher.md), [`jwtnone`](analyzers/jwtnone.md), [`weakpwhash`](analyzers/weakpwhash.md)) complement gosec.

## Example

A `.golangci.yml` that turns on tidygo and a few of the companions above:

```yaml
version: "2"

linters:
  enable:
    - tidygo
    - bodyclose
    - containedctx
    - gocognit
    - gocritic
    - gosec
  settings:
    custom:
      tidygo:
        type: "module"
    gocritic:
      enabled-checks:
        - paramTypeCombine
```
