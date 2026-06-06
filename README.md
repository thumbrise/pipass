# pipass

Compile-time type-safe pipeline pass wrappers generator for Go.

## What you get

You define:

```go
type Score int64

type Player struct {
    Name   string
    Score  Score
    Active *bool
}

type Session struct {
    Title    string
    Host     Player
    Players  []Player
    Metadata map[string]string
}
```

Generated API overview:

```
PlayerPass
    Name() string
    SetName(value string, reason string)
    Score() Score
    SetScore(value Score, reason string)
    Active() *bool
    SetActive(value *bool, reason string)

SessionPass
    Title() string
    SetTitle(value string, reason string)
    Host() PlayerPass
    SetHost(value PlayerPass, reason string)
    Players() []PlayerPass
    MapPlayers(fn func(PlayerPass) error) error
    AppendPlayers(value PlayerPass, reason string)
    MetadataKey(key string) string
    SetMetadataKey(key string, value string, reason string)
```

Every mutation flows through Ledger with path tracking, idempotency guard, and a human-readable reason. Named types like `Score` are preserved as-is.

## Description

pipass generates type-safe shadow wrappers (Pass/PipePass) for your Go structs. Every getter, setter, and mutation is wired to a **Ledger** that records what changed, where, and why — without touching your original types.

## Why

Go structs are silent. You set a field, it changes, and nobody knows. Logging every mutation manually is tedious and error-prone. Framework-level solutions are too heavy.

pipass gives you observability at the struct field level with zero runtime reflection, zero dependencies in your domain types, and full Go type safety.

## Use cases

- **Audit trails** — track every state change with human-readable reasons
- **Pluggable pipelines** — traverse and mutate deep object graphs with automatic path propagation
- **Multi-stage workflows** — connect distinct processing stages through a shared observable graph
- **Debugging** — replay or inspect every mutation in a complex data flow
- **State machines** — observe transitions with full history

## How

1. Define your structs
2. Register the types you want to observe with `pipass.Compile(outputPackage, rootType, nodeTypes...)`
3. It generates `TPass` (interface) and `TPipePass` (struct with ledger) for each type
4. Use the generated wrappers in your pipeline — every mutation flows through the ledger

## What

| Feature               | Example                                                                               |
|-----------------------|---------------------------------------------------------------------------------------|
| **Type preservation** | `*bool`, `*int`, `*string`, named types, `time.Time` — exact types, not `interface{}` |
| **Slice nodes**       | `[]Player` → `MapPlayers(fn)` / `AppendPlayers(value, reason)`                        |
| **Singular nodes**    | Any registered struct becomes an observable node with its own `_path`/`_ledger`       |
| **Maps**              | `map[string]string` → `Key(key) string` / `SetKey(key, value, reason)`                |
| **Ledger**            | Pluggable — built-in `PrintLedger` or your own implementation                         |
| **Path propagation**  | Automatic — `session.Players[0].Stats.HP` is tracked without manual path strings      |
| **Idempotency guard** | Setters skip logging when value doesn't change (`reflect.DeepEqual`)                  |
| **Go 1.26**           | Full support for `new(expr)` syntax for pointer fields                                |
| **No codec**          | Plain Go structs in, plain Go structs out                                             |

## Details

- The generated code is a **readable Go source file** — no bytecode, no plugins
- Types are analyzed at compile‑time via `reflect` — no runtime overhead
- Add the generated file to version control; regenerate when your DTOs change
- `pipass.Compile` takes zero‑value exemplars, not factory functions
- Maps are exposed via `Key(key)` / `SetKey(key, value, reason)` naming — the key type matches the original map key type (`string`, `int`, etc.)
- Mutations before a singular node is attached are not logged (pass a ledger to the constructor if needed)

## Examples

- [examples/game](examples/game) — full demo with pointers, slices, maps, singular nodes, and ledger
- [examples/poc](examples/poc) — minimal proof-of-concept with Person and Pet

## License

Apache 2.0
