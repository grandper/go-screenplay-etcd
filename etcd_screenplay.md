# Design Document: `screenplay` package (ETCD Screenplay extension)

## Status
This document is a complete, self-contained specification. An independent
engineer or agent should be able to re-implement the entire `screenplay/`
folder from scratch using only the information below, without access to the
original source code.

## 1. Purpose and philosophy

This package is an **extension** to a third-party, generic *Screenplay
pattern* library for Go (the "screenplay core library"). The Screenplay
pattern is a BDD-friendly way of writing acceptance tests and automation
code: instead of modelling the system under test, you model the **Actors**
(the users/personas of your tests) and describe what they can do in plain,
readable sentences such as:

> *Alice attempts to put the value "postgresql://localhost:5432" for the key
> "app/config/database".*
>
> *Alice sees that the value for the key "app/config/database" equals
> "postgresql://localhost:5432".*

The screenplay core library defines four kinds of building blocks that
extensions plug into:

| Concept | Core interface | Meaning |
|---|---|---|
| **Ability** | `interface { Forget() error }` | Something an actor *can do* technically (e.g. talk to a specific backend). Abilities are attached to an actor with `actor.WhoCan(...)` and released with `actor.Exit()`. |
| **Action / Task (`Performable`)** | `interface { PerformAs(actor *Actor) error; String() string }` | Something an actor *does*. Executed through `actor.AttemptsTo(...)`. |
| **Question** | `interface { AnsweredBy(actor *Actor) (any, error); String() string }` | Something an actor *asks about* the system state. Executed through `actor.AsksFor(...)` / `actor.Sees(...)`, or indirectly through the `see` action builder. |
| **Resolution** | `interface { Resolve() Matcher; String() string }` where `Matcher func(obj any) (bool, error)` | A reusable, composable assertion used to check the answer to a question (e.g. "equals X", "exists", "contains Y"). |

The library also exposes:
- `screenplay.ActorNamed(name string) *Actor` — creates an actor.
- `actor.WhoCan(abilities ...Ability) *Actor` — grants abilities.
- `screenplay.UseAbilityTo[A Ability]().Of(actor) (A, error)` — a generic
  helper that type-safely extracts a previously granted ability from an
  actor, returning an error (wrapping a sentinel "actor does not have the
  required ability" error) if the actor never learned it.
- `actor.AttemptsTo(actions ...Performable) error` — runs a sequence of
  actions/tasks, stopping at (and returning) the first error.
- `see.The(question).Is(resolution) Performable` (and the family of verbs
  `IsNot`, `Does`, `Has`, `Have`, ... — all meaning "assert that the answer to
  this question satisfies/violates this resolution") — a generic bridge that
  turns "question + resolution" into something you can pass to
  `AttemptsTo`.
- A family of ready-made resolutions such as `equal.To[T](value)` ("equals"),
  from which this extension's own `resolution.Exists()` takes inspiration.

**This extension's job** is to teach the Screenplay vocabulary how to talk to
an **etcd v3 key-value store**: one new Ability (`UseETCD`), a small set of
new Actions (`PutTheValue`, `DeleteKey`, `DeleteKeysStartingWith`), one new
Question (`ValueForKey` / `ValuesStartingWithPrefix`), and one new Resolution
(`Exists`).

The package purposefully has **zero knowledge of etcd wiring beyond the
minimal `Put`/`Get`/`Delete`/`Close` operations**; it does not implement
leases, watches, transactions, or distributed locks — those are out of
scope.

## 2. Module layout

Assume the new code lives in its own Go module (placeholder module path used
throughout this document: `example.org/go-screenplay-etcd`; replace with the
real module path when implementing). The package tree to (re)create is:

```
screenplay/
├── README.md                     # user-facing docs (see §8)
├── ability/
│   ├── etcd_client.go             # EtcdClient interface + real implementation
│   ├── etcd_client_test.go        # (optional) unit tests for the real wrapper
│   ├── connect_to_etcd.go         # UseETCDAbility + ETCDBuilder
│   ├── connect_to_etcd_test.go
│   └── testdata/
│       ├── valid_client.crt       # a self-signed cert usable for TLS tests
│       ├── valid_client.key
│       ├── invalid_client.crt     # deliberately malformed cert content
│       └── invalid_client.key
├── action/
│   ├── put_value.go
│   ├── put_value_test.go
│   ├── delete_key.go
│   └── delete_key_test.go
├── question/
│   ├── get_value.go
│   └── get_value_test.go
└── resolution/
    ├── exists.go
    └── exists_test.go
```

Additionally, at the **root of the module** (one level above `screenplay/`),
there must be a hand-written **test double** for the etcd client interface,
used by every unit test in the `screenplay/` tree:

```
mock_etcd_client.go   # package example.org/go-screenplay-etcd (root package)
```

This mock is a `testify/mock`-based stub implementing the same `EtcdClient`
interface defined in `screenplay/ability`. It lives at the module root
(rather than inside `screenplay/ability`) purely so that it can be imported
by `action`, `question`, and `ability` test files alike without creating
import cycles (those packages all depend on `ability`, so the mock cannot
live there if `action`/`question` need it too; putting it at the module root
sidesteps the cycle). Name the mock type `MockETCDClient`.

## 3. External dependencies

| Dependency | Why |
|---|---|
| Screenplay core library (`.../screenplay` package) | `Actor`, `Ability`, `Performable`, `Question`, `Resolution`, `Matcher`, `UseAbilityTo` |
| `go.etcd.io/etcd/client/v3` (`clientv3`) | The real etcd v3 client (`clientv3.Client`, `clientv3.Config`, `clientv3.OpOption`, `clientv3.WithPrefix()`, `clientv3.PutResponse`, `clientv3.GetResponse`, `clientv3.DeleteResponse`) |
| `go.etcd.io/etcd/api/v3/mvccpb` | `mvccpb.KeyValue`, used only in tests to build fake `GetResponse.Kvs` |
| `github.com/rs/zerolog/log` | Structured logging; used once, to log that a secure (TLS) connection was configured |
| `github.com/stretchr/testify` (`assert`, `require`, `mock`) | Assertions and the mocking framework used throughout |
| (optional) `github.com/cucumber/godog` | If Gherkin `.feature` files are added for black-box/BDD-style tests (see §9) |

## 4. Package `ability`

### 4.1 `EtcdClient` interface (file `etcd_client.go`)

Define a narrow interface that captures only the subset of the etcd v3 client
surface this extension needs. This is the seam that makes the package
mockable in tests, and that different "flavors" of connection (plain /
TLS) can share:

```go
package ability

import (
    "context"

    clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdClient defines the interface for an ETCD client.
type EtcdClient interface {
    Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
    Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
    Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
    Close() error
}
```

Provide a thin production implementation that simply forwards every call to
a real `*clientv3.Client`:

```go
// RealEtcdClient is a concrete implementation of EtcdClient using the real ETCD client.
type RealEtcdClient struct {
    Client *clientv3.Client
}

func (r *RealEtcdClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
    return r.Client.Put(ctx, key, val, opts...)
}

func (r *RealEtcdClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
    return r.Client.Get(ctx, key, opts...)
}

func (r *RealEtcdClient) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
    return r.Client.Delete(ctx, key, opts...)
}

func (r *RealEtcdClient) Close() error {
    return r.Client.Close()
}
```

### 4.2 `UseETCDAbility` and `ETCDBuilder` (file `connect_to_etcd.go`)

The ability itself only needs to hold the client:

```go
// UseETCDAbility represents the ability of an actor to use the ETCD client.
type UseETCDAbility struct {
    Client EtcdClient
}

// Forget cleans up resources used by the ETCD ability.
func (c *UseETCDAbility) Forget() error {
    if c.Client != nil {
        return c.Client.Close()
    }
    return nil
}
```

`Forget` is the hook the screenplay core library calls automatically when
`actor.Exit()` is invoked, and it **must be nil-safe**: an ability built from
a builder that failed to produce a client might be partially constructed by
callers who ignore the nil return, so `Forget` on a `nil` `Client` must
return `nil`, never panic.

Connections are configured through a fluent builder, because an actor needs
to be able to write a single expression such as:

```go
alice := screenplay.ActorNamed("alice").WhoCan(
    ability.UseETCD().WithUrls([]string{"localhost:2379"}),
)
```

Design the builder as follows:

```go
// ETCDBuilder helps to build UseETCDAbility with various configurations.
type ETCDBuilder struct {
    certFile string
    keyFile  string
    urls     []string
}

// UseETCD initializes an ETCDBuilder for configuring the ETCD ability.
func UseETCD() *ETCDBuilder {
    return &ETCDBuilder{
        urls:     []string{},
        certFile: "",
        keyFile:  "",
    }
}

// WithUrls configures the ETCD ability with the provided URLs without SSL.
func (b *ETCDBuilder) WithUrls(urls []string) *UseETCDAbility {
    b.urls = urls
    if etcd, err := b.makeETCDConnection(); err == nil {
        return etcd
    }
    return nil
}

// WithSSL configures the ETCD ability with mutual-TLS client certificates.
func (b *ETCDBuilder) WithSSL(urls []string, certFile string, keyFile string) *UseETCDAbility {
    b.certFile = certFile
    b.keyFile = keyFile
    b.urls = urls
    if etcd, err := b.makeETCDConnection(); err == nil {
        return etcd
    }
    return nil
}
```

**Important, deliberate API trait to preserve:** `WithUrls` and `WithSSL`
return `*UseETCDAbility`, **not** `(*UseETCDAbility, error)`. This keeps the
call sites in the README one-liner-friendly (`actor.WhoCan(ability.UseETCD().WithUrls(...))`),
at the cost of silently returning `nil` when the connection cannot be built
(for example, no URLs were given, or SSL files cannot be loaded). The real
`clientv3.New` rarely fails synchronously for unreachable endpoints (it dials
lazily), so this mostly matters for configuration errors, not network
errors. Implementers re-doing this from scratch must keep this nil-on-error
contract because it directly affects how calling code and tests are
expected to behave (see the test table in §7.1).

The actual connection-building logic is private and *does* return an error,
so it can be unit-tested precisely:

```go
func (b *ETCDBuilder) makeETCDConnection() (*UseETCDAbility, error) {
    if b == nil {
        return nil, errors.New("ETCDBuilder is nil")
    }
    if len(b.urls) == 0 {
        return nil, errors.New("urls cannot be nil or empty")
    }

    cfg := clientv3.Config{Endpoints: b.urls}

    if b.certFile != "" && b.keyFile != "" {
        tlsConfig, err := getTLSConfig(b.certFile, b.keyFile)
        if err != nil {
            return nil, err
        }
        cfg.TLS = tlsConfig
    }

    client, err := clientv3.New(cfg)
    if err != nil {
        return nil, err
    }

    return &UseETCDAbility{Client: &RealEtcdClient{Client: client}}, nil
}
```

Rules to respect:
- A `nil` receiver must be handled explicitly and produce the exact error
  message `"ETCDBuilder is nil"` (do not let it panic on a nil pointer
  dereference).
- Empty or `nil` `urls` must produce exactly `"urls cannot be nil or empty"`.
- TLS is only attempted when **both** `certFile` and `keyFile` are
  non-empty; if only one of the two is set, silently skip TLS and connect in
  plaintext (this mirrors `clientv3.New`'s tolerant behavior and is covered
  by tests, see §7.1).
- Any error returned by `getTLSConfig` (bad cert/key path or content) must
  propagate unchanged.
- `clientv3.New` itself practically never errors for syntactically valid
  (even unreachable) URLs, since etcd client v3 dials lazily; this means most
  "happy path" tests in this builder can use clearly-invalid-looking host
  strings (e.g. `"invalid-url"`) and still expect **no error** — only a
  non-functional client that would fail on first real RPC. Keep that
  characteristic in mind when writing the matching tests.

### 4.3 TLS loading helper

```go
var errFailedToLoadCert = errors.New("failed to load client key pair")

func getTLSConfig(certFile, keyFile string) (*tls.Config, error) {
    clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("%w: %w", errFailedToLoadCert, err)
    }

    log.Info().Msg("Etcd client certificate and key loaded. Using secure connection.")

    return &tls.Config{
        Certificates: []tls.Certificate{clientCert},
        MinVersion:   tls.VersionTLS12, // enforce a secure minimum TLS version
    }, nil
}
```

Rules:
- Any failure of `tls.LoadX509KeyPair` (missing file, malformed PEM,
  mismatched cert/key pair, etc.) must be wrapped with `errFailedToLoadCert`
  so callers/tests can assert with `errors.Is(err, errFailedToLoadCert)`.
- `MinVersion` must be pinned to TLS 1.2 or higher (this is a security
  requirement, not just a style choice — it avoids static-analysis
  "TLS MinVersion too low" findings).
- Log a single informational line confirming a secure connection was
  configured; do not log the certificate or key contents.

### 4.4 Interface compliance assertions

At the bottom of `connect_to_etcd.go`, assert the implemented interface at
compile time:

```go
var _ screenplay.Ability = (*UseETCDAbility)(nil)
```

## 5. Package `action`

Both actions in this package follow the same shape: a small struct holding
the parameters, a constructor (or pair of constructors) with a fluent,
English-sentence-like name, a `PerformAs` method that fetches the
`UseETCDAbility` from the actor and calls the etcd client, and a `String`
method used for narration/logging.

### 5.1 `PutTheValue` / `ForTheKey` (file `put_value.go`)

```go
// PutValueAction represents an action to put a key-value pair into ETCD.
type PutValueAction struct {
    Key   string
    Value string
}

// PutTheValue creates an action to put a value into ETCD. The key is set
// separately via ForTheKey, so the call reads like a sentence:
//   action.PutTheValue("postgresql://localhost:5432").ForTheKey("app/config/database")
func PutTheValue(value string) *PutValueAction {
    return &PutValueAction{Value: value}
}

// ForTheKey sets the key for the PutValueAction.
func (p *PutValueAction) ForTheKey(key string) *PutValueAction {
    p.Key = key
    return p
}

// PerformAs executes the action as the given actor.
func (p *PutValueAction) PerformAs(actor *screenplay.Actor) error {
    etcd, err := screenplay.UseAbilityTo[*ability.UseETCDAbility]().Of(actor)
    if err != nil {
        return err
    }
    _, err = etcd.Client.Put(context.Background(), p.Key, p.Value)
    return err
}

// String describes the action.
func (p *PutValueAction) String() string {
    return fmt.Sprintf("put value '%s' for key '%s'", p.Value, p.Key)
}

var _ screenplay.Performable = (*PutValueAction)(nil)
```

Rules:
- `PerformAs` must return the ability-lookup error verbatim (do not wrap it)
  when the actor never learned `UseETCDAbility`.
- `PerformAs` must return the raw error from `Client.Put` verbatim (no
  wrapping) on failure.
- `Put` is always called **without** any `clientv3.OpOption` (no TTL/lease
  support in this version).
- Calling `PerformAs` with a `nil` actor must panic (this is standard
  behavior inherited from the screenplay core library, not something this
  package needs to special-case; just don't add defensive nil-checks that
  would turn the panic into a silent no-op).

### 5.2 `DeleteKey` / `DeleteKeysStartingWith` (file `delete_key.go`)

```go
// DeleteKeyAction represents an action to delete a key (or a set of keys
// sharing a prefix) from ETCD.
type DeleteKeyAction struct {
    Key        string
    WithPrefix bool
}

// DeleteKey creates an action to delete a specific key.
func DeleteKey(key string) *DeleteKeyAction {
    return &DeleteKeyAction{Key: key, WithPrefix: false}
}

// DeleteKeysStartingWith creates an action to delete all keys starting with the given prefix.
func DeleteKeysStartingWith(prefix string) *DeleteKeyAction {
    return &DeleteKeyAction{Key: prefix, WithPrefix: true}
}

// PerformAs executes the action as the given actor.
func (d *DeleteKeyAction) PerformAs(actor *screenplay.Actor) error {
    etcd, err := screenplay.UseAbilityTo[*ability.UseETCDAbility]().Of(actor)
    if err != nil {
        return err
    }

    if d.WithPrefix {
        _, err = etcd.Client.Delete(context.Background(), d.Key, clientv3.WithPrefix())
    } else {
        _, err = etcd.Client.Delete(context.Background(), d.Key)
    }
    return err
}

// String describes the action.
func (d *DeleteKeyAction) String() string {
    if d.WithPrefix {
        return "delete keys starting with prefix '" + d.Key + "'"
    }
    return "delete key '" + d.Key + "'"
}

var _ screenplay.Performable = (*DeleteKeyAction)(nil)
```

Rules:
- When `WithPrefix` is `false`, `Delete` must be called with **zero**
  `clientv3.OpOption` arguments (an empty variadic slice) — not a slice
  containing a "no-op" option. Tests assert on `len(opts) == 0` in that
  case.
- When `WithPrefix` is `true`, `Delete` must be called with exactly one
  option, `clientv3.WithPrefix()`.
- Error propagation and nil-actor panic rules are the same as for
  `PutValueAction`.

## 6. Package `question`

### 6.1 `ValueForKey` / `ValuesStartingWithPrefix` (file `get_value.go`)

```go
// ValueOfQuestion represents a question to get the value of a key or keys from ETCD.
type ValueOfQuestion struct {
    key        string
    withPrefix bool
}

// ValueForKey creates a question to get the value of a specific key.
func ValueForKey(key string) *ValueOfQuestion {
    return &ValueOfQuestion{key: key, withPrefix: false}
}

// ValuesStartingWithPrefix creates a question to get all values whose key starts with prefix.
func ValuesStartingWithPrefix(prefix string) *ValueOfQuestion {
    return &ValueOfQuestion{key: prefix, withPrefix: true}
}

// AnsweredBy executes the question as the given actor.
func (q *ValueOfQuestion) AnsweredBy(actor *screenplay.Actor) (any, error) {
    etcd, err := screenplay.UseAbilityTo[*ability.UseETCDAbility]().Of(actor)
    if err != nil {
        return nil, err
    }

    resp, err := q.getValuesFromETCD(etcd)
    if err != nil {
        return nil, err
    }

    result := make(map[string]string)
    for _, kv := range resp.Kvs {
        result[string(kv.Key)] = string(kv.Value)
    }

    return result, nil
}

func (q *ValueOfQuestion) getValuesFromETCD(etcd *ability.UseETCDAbility) (*clientv3.GetResponse, error) {
    if q.withPrefix {
        return etcd.Client.Get(context.Background(), q.key, clientv3.WithPrefix())
    }
    return etcd.Client.Get(context.Background(), q.key)
}

// String describes the question.
func (q *ValueOfQuestion) String() string {
    return fmt.Sprintf("content of the key named '%s' and with prefix '%t'", q.key, q.withPrefix)
}

var _ screenplay.Question = (*ValueOfQuestion)(nil)
```

Rules — these are the behaviors most heavily covered by tests, get them
exactly right:
- The answer type is always `map[string]string` (declared as `any` to
  satisfy the `Question` interface), **never** `nil`, even when zero keys
  matched: an empty `Get` response must produce an **empty, non-nil map**
  (`map[string]string{}`), not a `nil` map. This matters because
  `resolution.Exists()` (see §7) treats a non-nil empty map as "exists".
- On any error from the ability lookup or from `Client.Get`, return
  `(nil, err)` — the first return value must be `nil`, not an empty map,
  when there is an error.
- `Get` must be called **without** options for a plain key lookup, and with
  exactly `clientv3.WithPrefix()` for a prefix lookup — mirroring the
  `DeleteKeyAction` convention.
- Keys and values in the result map come straight from
  `string(kv.Key)` / `string(kv.Value)`; don't trim, decode, or otherwise
  transform them.

## 7. Package `resolution`

### 7.1 `Exists` (file `exists.go`)

```go
// ExistsResolution checks if a key exists (is not nil).
type ExistsResolution struct{}

// Exists creates a resolution to check if a key exists.
func Exists() *ExistsResolution {
    return &ExistsResolution{}
}

// Resolve returns a matcher that checks if the value is not nil.
func (r *ExistsResolution) Resolve() screenplay.Matcher {
    return func(obj any) (bool, error) {
        return obj != nil, nil
    }
}

// String describes the resolution's expectation.
func (r *ExistsResolution) String() string {
    return "exists"
}

var _ screenplay.Resolution = (*ExistsResolution)(nil)
```

Rules and nuance to preserve exactly (covered extensively by tests):
- The matcher never returns an error; its second return value is always
  `nil`.
- The check is a plain `obj != nil` comparison on the `any` wrapper. This
  has a well-known Go gotcha that must be preserved intentionally, not
  "fixed": a **typed nil pointer** boxed into an `any` (e.g.
  `any((*string)(nil))`) is **not** equal to untyped `nil`, so
  `Exists().Resolve()` returns `true` for it. Likewise it returns `true` for
  zero values that are *not* nil, such as `0`, `false`, `""`, an empty slice,
  or an empty (but non-nil) map — only a bare untyped `nil` makes it return
  `false`.
- `Resolve()` must return a brand-new closure each time it is called (no
  shared mutable state), and repeated invocations of the same returned
  matcher with the same input must be idempotent.

This resolution is what lets `question.ValueForKey`/`ValuesStartingWithPrefix`
answers (always a non-nil map, per §6.1) be asserted against with
`.Is(resolution.Exists())` to mean "the key (or at least one key under the
prefix) was found" — note that because an *empty* map is still non-nil, this
particular resolution only reliably detects "the question could be
answered", while actually distinguishing "zero matches" from "one or more
matches" requires a different resolution (e.g. a future "is not empty" or
"has length greater than 0" resolution) — call this out explicitly as a
known limitation in the README rather than silently special-casing empty
maps inside `Exists`.

## 8. `README.md` content and philosophy

The README is the single most important artifact for discoverability: it
must show, in order:
1. **New Abilities** — plain vs TLS connection, one-liner each.
2. **New Actions** — storing a value, deleting a single key, deleting by
   prefix.
3. **New Questions** — retrieving a value, retrieving by prefix, checking
   existence. Use the **real, verified-correct call shape**:
   `actor.AttemptsTo(see.The(question.ValueForKey("app/config/database")).Is(equal.To("postgresql://localhost:5432")))`
   and
   `actor.AttemptsTo(see.The(question.Key("app/config/")).Is(resolution.Exists()))`.
   (Do not reproduce `see.The(x, y)` two-argument call shapes or bare
   `err := anActor.AttemptsTo(see.The(...)), is.EqualTo(...)` — these are
   invalid Go and must not appear in the final README; `see.The` takes a
   single `Question` argument and returns an `*Assertion` on which you chain
   `.Is(...)`/`.IsNot(...)`/etc.)
4. **A complete, compilable example** `package main` that: creates an actor,
   grants it the ETCD ability, stores two values, asserts existence and
   equality, deletes a key and a prefix, and calls `actor.Exit()` at the end
   to release the connection. The example must build without a duplicated
   code block and without unused imports — every import listed must be used
   exactly as shown.
5. **Installation** instructions (`go get <module path>`).

Keep the tone and structure of the existing README (headings, short
intros, fenced `go` snippets) but fix the two defects present in the
original: (a) a copy-pasted duplicate "Store configuration values" block in
the complete example, and (b) the non-existent `is.EqualTo(...)` /
`see.The(a, b)` call shapes described above.

## 9. Test case catalog

Re-implement, at minimum, the following behaviors as table-driven
`testing.T` unit tests (using `testify/assert`, `testify/require`, and
`testify/mock`). Use a `//nolint:testpackage` same-package test style (tests
live in `package ability`, `package action`, etc., not `_test` suffixed
packages) so that private fields (`key`, `withPrefix`, `urls`, `certFile`,
`keyFile`) can be asserted on directly.

### 9.1 `ability` package

**`UseETCD()`**
- Returns a non-nil `*ETCDBuilder` with empty `urls`, `certFile`, `keyFile`.

**`UseETCDAbility.Forget()`**
- Closing succeeds → returns `nil`, and the mock's `Close` expectation is
  satisfied.
- Closing fails → returns the underlying error unchanged.
- `Client` is `nil` → returns `nil` without panicking.

**`ETCDBuilder.WithUrls(urls)`**
- Empty slice → returns `nil`; builder's `urls` field is still set to the
  (empty) input.
- `nil` slice → returns `nil`; builder's `urls` field remains `nil`.
- Non-empty slice (even a syntactically-invalid host like `"invalid-url"`)
  → returns a non-nil `*UseETCDAbility`; builder's `urls` reflects the
  input.

**`ETCDBuilder.WithSSL(urls, certFile, keyFile)`**
- Empty `urls` → returns `nil` even though cert/key fields got recorded.
- Non-empty `urls` with a cert/key path that doesn't exist on disk →
  returns `nil` (because TLS loading fails inside `makeETCDConnection`).
- Non-empty `urls` with empty `certFile`/`keyFile` (SSL effectively
  disabled) → returns a non-nil ability.
- In all cases, assert the builder's `urls`, `certFile`, `keyFile` fields
  were set to exactly the given arguments (even when the result is `nil`),
  since `WithUrls`/`WithSSL` mutate the receiver before deciding what to
  return.

**`ETCDBuilder.makeETCDConnection()`**
- `nil` receiver → error `"ETCDBuilder is nil"`, result `nil`.
- `nil` urls → error `"urls cannot be nil or empty"`.
- empty urls slice → same error.
- non-empty, "invalid-looking" URL and no TLS → **no error**, non-nil
  result (because `clientv3.New` dials lazily).
- non-empty URL with a cert/key path that doesn't exist → error (TLS load
  failure propagated), nil result.
- non-empty URL with only `certFile` set (no `keyFile`) → **no error**
  (TLS block skipped because both must be non-empty).
- non-empty URL with only `keyFile` set (no `certFile`) → **no error**,
  same reasoning.
- multiple URLs → accepted and stored verbatim on the builder.

**`getTLSConfig(certFile, keyFile)`**
- Nonexistent cert/key paths → error that `errors.Is(err, errFailedToLoadCert)`.
- Empty cert path (with a non-empty but nonexistent key path) → error.
- Empty key path → error.
- Both empty → error.
- A valid cert/key pair (use the `testdata/valid_client.crt` /
  `valid_client.key` fixtures) → non-nil `*tls.Config`, `MinVersion` equals
  `tls.VersionTLS12`, `Certificates` is non-empty.
- An invalid/malformed cert or key pair (use `testdata/invalid_client.crt` /
  `invalid_client.key`) → error wrapping `errFailedToLoadCert`.

**Chaining / state behavior**
- Calling `WithUrls` then `WithSSL` on the same builder updates `urls` to the
  second call's value and sets `certFile`/`keyFile` accordingly (builder
  state is mutable and sequential, not accumulative/additive).

### 9.2 `action` package — `PutValueAction`

- Successful put with populated key/value → no error; mock's `Put` is
  called with `(ctx, "the-key", "the-value", <empty opts>)`.
- Actor without the `UseETCDAbility` → returns the ability-lookup error.
- Underlying `Put` failure → returns that error unchanged.
- Empty key, empty value, or both empty → still forwarded as-is to `Put`,
  no special-casing, no error.
- Keys/values containing special characters (`/`, `-`, `_`, `.`, spaces,
  symbols like `&`, `!`, `@`, `#`), JSON-shaped strings, and Unicode
  (including multi-byte runes and emoji) must be passed through unchanged.
- `PerformAs(nil)` panics.
- Field-only tests: constructing `&PutValueAction{...}` directly and reading
  back `Key`/`Value`, including very long strings.

### 9.3 `action` package — `DeleteKeyAction`

- `DeleteKey(key)` / `DeleteKeysStartingWith(prefix)` constructors set
  `WithPrefix` to `false` / `true` respectively.
- Successful delete (no prefix) → `Delete` called with zero extra options
  (`len(opts) == 0`); no error.
- Successful delete (with prefix) → `Delete` called with exactly one option
  (`len(opts) > 0`, in practice exactly 1); no error.
- Underlying `Delete` failure (both prefixed and non-prefixed) → error
  propagated unchanged.
- Actor missing the ability → ability-lookup error returned, `Delete` never
  called.
- Empty key with/without prefix → still forwarded, no error, no special
  casing.
- Special characters (e.g. `"app/config/database"`) work the same with or
  without prefix.
- `PerformAs(nil)` panics.
- Field validation: default `WithPrefix` is `false` when constructing
  `&DeleteKeyAction{Key: "..."}` without explicitly setting it.

### 9.4 `question` package — `ValueOfQuestion`

- `ValueForKey(key)` sets `key` and `withPrefix=false`; works for empty key
  and for keys containing special characters.
- `ValuesStartingWithPrefix(prefix)` sets `key=prefix` and
  `withPrefix=true`; works for empty prefix and complex/nested path
  prefixes.
- `AnsweredBy`:
  - Single existing key → returns `map[string]string{key: value}`.
  - Single non-existing key (empty `Kvs`) → returns `map[string]string{}`
    (empty, non-nil), no error.
  - Prefix search matching several keys → returns a map with one entry per
    returned `Kvs` item, keyed by the raw returned key bytes as a string.
  - Prefix search with zero matches → empty, non-nil map.
  - Actor missing the ability → `(nil, err)`.
  - `Get` failure (both prefixed and non-prefixed) → `(nil, err)`, error
    message passed through unchanged.
  - Empty key string (non-prefixed) and empty prefix string (prefixed) are
    valid inputs and must be forwarded to `Get` as-is.
  - Unicode keys/values (e.g. CJK characters) round-trip correctly.
  - A key whose value is the empty string round-trips to
    `map[string]string{key: ""}`, not omitted from the map.
  - `AnsweredBy(nil)` panics.
- Field-only construction tests mirroring the ability tests above (long
  keys, path-like keys combined with `withPrefix=true`).

### 9.5 `resolution` package — `ExistsResolution`

- `Resolve()` applied to: a non-nil string (incl. empty string `""`), `0`,
  `true`, `false`, a non-nil map (incl. empty map), a non-nil slice (incl.
  empty slice), a struct value, a pointer to a struct, a non-nil `error`
  value, and a typed nil pointer boxed in `any` (`(*string)(nil)`) → all
  return `(true, nil)`.
- `Resolve()` applied to untyped `nil` → returns `(false, nil)`.
- `Exists().String()` returns exactly `"exists"`.
- Calling the same matcher multiple times with the same input is
  idempotent and returns the same result every time.
- Two separate calls to `Resolve()` on the same (or equivalent)
  `ExistsResolution` each produce independently-usable matchers that behave
  consistently given their respective inputs.

### 9.6 Cross-package / black-box scenarios (optional but recommended)

If Gherkin/BDD-style verification is desired on top of the Go unit tests
above (the surrounding project already depends on a Gherkin-running test
framework), model scenarios such as the ones below, driven against a real
or embedded etcd server, with step definitions that use the actors,
actions, and questions exactly as documented in the README:

```gherkin
Feature: Screenplay actors operate an ETCD store

  Background:
    Given a running ETCD server
    And an actor named "Alice" who can use ETCD

  Scenario: Storing and retrieving a value
    When Alice attempts to put the value "postgresql://localhost:5432" for the key "app/config/database"
    Then Alice sees that the value for the key "app/config/database" is "postgresql://localhost:5432"

  Scenario: Checking that a key exists
    Given Alice has put the value "redis://localhost:6379" for the key "app/config/cache"
    Then Alice sees that the key "app/config/" exists

  Scenario: Deleting a single key
    Given Alice has put the value "redis://localhost:6379" for the key "app/config/cache"
    When Alice attempts to delete the key "app/config/cache"
    Then Alice sees that the value for the key "app/config/cache" is an empty map

  Scenario: Deleting keys by prefix
    Given Alice has put the value "postgresql://localhost:5432" for the key "app/config/database"
    And Alice has put the value "redis://localhost:6379" for the key "app/config/cache"
    When Alice attempts to delete the keys starting with "app/config/"
    Then Alice sees that the values starting with "app/config/" is an empty map

  Scenario: Reading a missing key returns no error and an empty result
    When Alice attempts to read the value for the key "does/not/exist"
    Then Alice sees that the value for the key "does/not/exist" is an empty map

  Scenario: An actor without the ETCD ability cannot perform ETCD actions
    Given an actor named "Bob" with no abilities
    When Bob attempts to put the value "x" for the key "y"
    Then an error is returned mentioning the missing ability
```

## 10. The `MockETCDClient` test double (module root)

Implement a `testify/mock`-based stub of `ability.EtcdClient`. Two
asymmetric call-recording conventions must be reproduced exactly, because
the existing tests' `mock.On(...)` matcher arguments depend on them:

```go
package <rootpackage>

import (
    "context"

    "github.com/stretchr/testify/mock"
    clientv3 "go.etcd.io/etcd/client/v3"
)

type MockETCDClient struct {
    mock.Mock
}

// Delete and Put record their variadic opts as a single slice argument,
// so tests match them with `mock.Anything` (one placeholder for the whole
// slice) rather than one placeholder per option.
func (m *MockETCDClient) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
    args := m.Called(ctx, key, opts)
    if resp, ok := args.Get(0).(*clientv3.DeleteResponse); ok {
        return resp, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockETCDClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
    args := m.Called(ctx, key, val, opts)
    if resp, ok := args.Get(0).(*clientv3.PutResponse); ok {
        return resp, args.Error(1)
    }
    return nil, args.Error(1)
}

// Get instead flattens each option into its own positional argument, so
// tests match each option individually (e.g. with
// mock.AnythingOfType("clientv3.OpOption")) rather than matching the slice
// as a whole.
func (m *MockETCDClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
    callArgs := []interface{}{ctx, key}
    for _, opt := range opts {
        callArgs = append(callArgs, opt)
    }

    args := m.Called(callArgs...)
    if resp, ok := args.Get(0).(*clientv3.GetResponse); ok {
        return resp, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *MockETCDClient) Close() error {
    args := m.Called()
    return args.Error(0)
}
```

Keep this inconsistency (slice-arg for `Put`/`Delete`, flattened args for
`Get`) intentionally — it is relied upon by the test suites in §9.2–9.4 and
changing it would require rewriting every `mock.On(...)` call across the
package.

## 11. Non-goals / explicit limitations to document in the README

- No support for leases/TTLs, watches, transactions, or distributed locks.
- No retry/backoff logic around etcd calls — failures surface immediately.
- `resolution.Exists()` only detects "non-nil answer", not "non-empty
  result set"; an empty map still "exists".
- Builders (`WithUrls`, `WithSSL`) swallow configuration errors into a
  `nil` ability rather than surfacing an error value, by design (see
  §4.2) — callers should treat a `nil` ability from `UseETCD()...` as a
  fatal misconfiguration and check it before calling `actor.WhoCan(...)`.

# Enrichment Notes: additional Abilities/Actions/Questions/Resolutions sourced from `go.etcd.io/etcd/client/v3` and `go.etcd.io/etcd/tests/v3`

## Status
This document is a companion to `screenplay_design.md`. It does **not**
replace or contradict that design; it proposes *additional* screenplay
vocabulary, derived from scanning the README and source of the two etcd
modules pinned in `go.mod`:

- `go.etcd.io/etcd/client/v3 v3.6.4` — the official etcd v3 Go client
  (`README.md`, `kv.go`, `op.go`, `lease.go`, `watch.go`, `txn.go`,
  `compare.go`, `maintenance.go`, `clientv3util/util.go`,
  `concurrency/{mutex,election,session}.go`, `namespace/`).
- `go.etcd.io/etcd/tests/v3 v3.6.4` — the etcd integration/e2e test
  framework (`integration/clientv3/examples/*_test.go`,
  `integration/clientv3/concurrency/example_*_test.go`).

Each finding below lists its source, why it is worth adding, the proposed
contract, and a Go snippet showing how an actor would use it, following the
same placeholder module path (`example.org/go-screenplay-etcd`) and
conventions established in `screenplay_design.md` (builders, `PerformAs`,
`AnsweredBy`, `Resolve`, nil-safety rules, etc.).

---

## 1. Time-to-live keys via Lease (`lease.go`)

**Source:** `go.etcd.io/etcd/client/v3/lease.go` defines the `Lease`
interface (`Grant`, `Revoke`, `TimeToLive`, `Leases`, `KeepAlive`,
`KeepAliveOnce`), and `op.go` defines `WithLease(leaseID LeaseID) OpOption`
for attaching a lease to a `Put`.

**Why it matters:** The current design only supports permanent keys. A very
common configuration/coordination use case is "this key should
self-expire after N seconds if nobody refreshes it" (feature flags,
heartbeats, soft locks, session markers). None of the existing actions
expose this.

**Proposed contract:**
- Extend the `EtcdClient` interface (in `ability`) with a `Lease` accessor,
  or add a dedicated narrow interface:
  ```go
  type LeaseClient interface {
      Grant(ctx context.Context, ttlSeconds int64) (*clientv3.LeaseGrantResponse, error)
      Revoke(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error)
  }
  ```
- New action `action.PutTheValueForDuration(value string) *PutValueWithTTLAction`,
  chainable with `.ForTheKey(key string)` and `.ThatExpiresAfter(ttl time.Duration)`.
  `PerformAs` first calls `Grant`, then `Put` with `clientv3.WithLease(leaseID)`.
- New question `question.TimeToLiveOfLease(leaseID clientv3.LeaseID)` wrapping
  `TimeToLive`, answering the remaining TTL in seconds (`int64`), `-1` meaning
  expired/unknown lease.

**Example usage:**

```go
alice := screenplay.ActorNamed("alice").WhoCan(
    ability.UseETCD().WithUrls([]string{"localhost:2379"}),
)

err := alice.AttemptsTo(
    action.PutTheValueForDuration("acquired").
        ForTheKey("locks/batch-job").
        ThatExpiresAfter(10 * time.Second),
)
if err != nil {
    log.Fatalf("failed to put a self-expiring key: %v", err)
}

// later, assert the lease is still alive and has some TTL left
err = alice.AttemptsTo(
    see.The(question.ValueForKey("locks/batch-job")).Is(resolution.Exists()),
)
```

---

## 2. Keeping a lease alive (`lease.go` — `KeepAlive`)

**Source:** `Lease.KeepAlive(ctx, id) (<-chan *LeaseKeepAliveResponse, error)`
in `lease.go`.

**Why it matters:** Pairs naturally with Finding 1. Long-running actors
(e.g. a worker holding a soft lock) need to periodically renew their lease
instead of letting it expire. This is the etcd-idiomatic alternative to
manually re-`Put`-ting a key on a timer.

**Proposed contract:**
- New action `action.KeepTheLeaseAlive(leaseID clientv3.LeaseID) *KeepLeaseAliveAction`.
  `PerformAs` starts the keep-alive channel and registers a background
  goroutine that drains it (logging/discarding responses), and stores a
  `context.CancelFunc` on the actor's memory (`actor.Remember("lease:<id>:cancel", cancel)`)
  so a matching cleanup task can stop it.
- A companion cleanup action `action.StopKeepingTheLeaseAlive(leaseID clientv3.LeaseID)`
  that recalls the cancel func and calls it; register it as an actor
  cleanup task (`actor.HasIndependentCleanupTasks(...)`) so lease renewal
  always stops when the actor exits the stage.

**Example usage:**

```go
leaseID := /* obtained from PutTheValueForDuration's answer, or a dedicated question.LeaseIDFor(key) */

err := alice.AttemptsTo(action.KeepTheLeaseAlive(leaseID))
if err != nil {
    log.Fatalf("failed to start keep-alive: %v", err)
}

alice.HasIndependentCleanupTasks(action.StopKeepingTheLeaseAlive(leaseID))
// ... alice.Exit() will now stop the keep-alive goroutine automatically.
```

---

## 3. Watching for changes (`watch.go`)

**Source:** `Watcher.Watch(ctx, key, opts...) WatchChan` in `watch.go`,
illustrated by `tests/v3/integration/clientv3/examples/example_watch_test.go`.

**Why it matters:** The existing `question.ValueForKey` is a one-shot
poll. Many acceptance scenarios are naturally event-driven: "when another
process updates this configuration key, the actor should notice it without
busy-polling". Watch is the etcd-native way to express that.

**Proposed contract:**
- New action `action.WaitUntilTheKeyChanges(key string) *WaitForChangeAction`,
  configurable with `.WithTimeout(d time.Duration)` (defaulting to the
  actor's `Timeout()`, see `screenplay.Actor.Timeout()`). `PerformAs` opens
  a `Watch(ctx, key)`, blocks until the first event or the timeout/context
  deadline, and stores the observed `clientv3.Event` in the actor's memory
  under a documented key (e.g. `"watch:"+key`) via `actor.Remember`, so a
  subsequent question can inspect it.
- New question `question.LatestEventOnKey(key string)` that simply recalls
  that memory slot (`actor.Recall("watch:" + key)`), answering `nil` if no
  watch ever completed for that key.
- A small helper resolution could live alongside: `resolution.WasPut()` /
  `resolution.WasDeleted()`, resolving `clientv3.Event.Type` against
  `clientv3.EventTypePut` / `clientv3.EventTypeDelete`.

**Example usage:**

```go
err := alice.AttemptsTo(
    action.WaitUntilTheKeyChanges("app/config/database").WithTimeout(5 * time.Second),
    see.The(question.LatestEventOnKey("app/config/database")).Is(resolution.WasPut()),
)
if err != nil {
    log.Fatalf("expected a config change notification: %v", err)
}
```

This directly supports Gherkin scenarios of the shape:

```gherkin
Scenario: An actor is notified when a configuration value changes
  Given Alice is watching the key "app/config/database"
  When another process updates the key "app/config/database"
  Then Alice observes a PUT event on the key "app/config/database"
```

---

## 4. Conditional writes / optimistic concurrency (`txn.go`, `compare.go`, `clientv3util/util.go`)

**Source:** `Txn.If(...).Then(...).Else(...).Commit()` in `txn.go`;
`Compare`, `Value`, `Version`, `ModRevision`, `CreateRevision` in
`compare.go`; ready-made helpers `clientv3util.KeyExists` /
`clientv3util.KeyMissing` in `clientv3util/util.go` (and their
`ExampleKeyMissing`/`ExampleKeyExists` tests).

**Why it matters:** `action.PutTheValue` always overwrites unconditionally.
A very common requirement — "only initialize this key if it doesn't
already exist" (avoiding clobbering another writer) or "only update this
key if nobody else changed it since I last read it" (optimistic locking) —
needs a transaction, not a plain `Put`.

**Proposed contract:**
- New action `action.PutTheValueIfAbsent(value string) *PutIfAbsentAction`,
  chainable with `.ForTheKey(key string)`. `PerformAs` runs:
  ```go
  etcd.Client.Txn(ctx).
      If(clientv3util.KeyMissing(key)).
      Then(clientv3.OpPut(key, value)).
      Commit()
  ```
  and turns a `!resp.Succeeded` outcome into a sentinel error
  `ErrKeyAlreadyExists` so callers/tests can `errors.Is` it.
- New action `action.PutTheValueIfUnchangedSince(value string, modRevision int64) *CompareAndSwapAction`,
  chainable with `.ForTheKey(key string)`, running:
  ```go
  etcd.Client.Txn(ctx).
      If(clientv3.Compare(clientv3.ModRevision(key), "=", modRevision)).
      Then(clientv3.OpPut(key, value)).
      Commit()
  ```
  returning a sentinel `ErrConcurrentModification` when `!resp.Succeeded`.
- This requires the `question.ValueForKey` family to also optionally expose
  the `ModRevision` (see Finding 6) so a caller can obtain the revision to
  compare against.

**Example usage:**

```go
err := alice.AttemptsTo(action.PutTheValueIfAbsent("initial-value").ForTheKey("app/config/flag"))
if errors.Is(err, action.ErrKeyAlreadyExists) {
    // another actor/process already initialized the flag; that's fine.
} else if err != nil {
    log.Fatalf("unexpected error initializing flag: %v", err)
}

// optimistic update pattern
meta, _ := alice.AsksFor(question.MetadataOfKey("app/config/flag"))
rev := meta.(question.KeyMetadata).ModRevision

err = alice.AttemptsTo(
    action.PutTheValueIfUnchangedSince("updated-value", rev).ForTheKey("app/config/flag"),
)
if errors.Is(err, action.ErrConcurrentModification) {
    log.Println("someone else updated the flag first, retry the read-modify-write")
}
```

---

## 5. Precise "at least one match" check via `WithCountOnly` (`op.go`)

**Source:** `WithCountOnly() OpOption` in `op.go`; `GetResponse.Count`
field (`RangeResponse.Count` in `etcdserverpb`), documented as "unaffected
by limits ... reflects the full count within the specified range".

**Why it matters:** `screenplay_design.md` §7/§11 explicitly documents a
known limitation: `resolution.Exists()` only detects "the question could be
answered" because `question.ValueOfQuestion` always answers a non-nil map,
even when zero keys matched. This etcd feature gives a cheap, accurate,
value-free way to answer "how many keys match?", which cleanly fixes that
gap without changing `ValueOfQuestion`'s existing (and already
test-covered) behavior.

**Proposed contract:**
- New question `question.NumberOfKeysMatching(prefix string) *CountQuestion`
  (always implies `WithPrefix()` + `WithCountOnly()`), answering a plain
  `int64` (`resp.Count`).
- No new resolution is required: the core screenplay library's
  `resolution/greater` package already provides `greater.Than(0)`, which is
  exactly what is needed to assert "one or more keys exist under this
  prefix" unambiguously — succeeding where a bare `resolution.Exists()`
  would be misleading on an empty result set.

**Example usage:**

```go
import "github.com/grandper/go-screenplay/resolution/greater"

err := alice.AttemptsTo(
    see.The(question.NumberOfKeysMatching("app/config/")).Is(greater.Than(int64(0))),
)
if err != nil {
    log.Fatalf("expected at least one key under app/config/: %v", err)
}
```

---

## 6. Key metadata / revision question (`kv.go`, `mvccpb.KeyValue`)

**Source:** `mvccpb.KeyValue` (via `go.etcd.io/etcd/api/v3/mvccpb`) carries
`CreateRevision`, `ModRevision`, `Version`, and `Lease` alongside `Key`/
`Value`; `compare.go` exposes matching `Cmp` builders
(`Version`, `CreateRevision`, `ModRevision`).

**Why it matters:** This single addition both (a) feeds Finding 4's
compare-and-swap action with the revision it needs, and (b) directly
enables the "Read and write using versioning" scenario already sketched in
the repository's existing `.feature` files (`... the version is
incremented`), which the current `screenplay` package has no way to assert
on at all — `ValueOfQuestion` only ever exposes the raw string value.

**Proposed contract:**
```go
// KeyMetadata is the answer type of question.MetadataOfKey.
type KeyMetadata struct {
    Value          string
    Version        int64
    CreateRevision int64
    ModRevision    int64
    Lease          int64
}

// MetadataOfKey creates a question for the full metadata of a single key.
func MetadataOfKey(key string) *MetadataQuestion
```
`AnsweredBy` performs a plain `Get(ctx, key)` and, when `len(resp.Kvs) == 0`,
answers the zero-value `KeyMetadata{}` paired with a `false` "found" flag
(or simply `nil`, mirroring how `ValueOfQuestion` treats "not found" — pick
one and document it, for consistency prefer returning `(nil, nil)`).

**Example usage:**

```go
answer, err := alice.AsksFor(question.MetadataOfKey("app/config/database"))
if err != nil {
    log.Fatalf("failed to read key metadata: %v", err)
}

meta := answer.(*question.KeyMetadata)
fmt.Printf("version=%d createRevision=%d modRevision=%d\n",
    meta.Version, meta.CreateRevision, meta.ModRevision)

err = alice.AttemptsTo(action.PutTheValue("postgresql://new-host:5432").ForTheKey("app/config/database"))
if err != nil {
    log.Fatalf("failed to update value: %v", err)
}

updated, _ := alice.AsksFor(question.MetadataOfKey("app/config/database"))
err = alice.AttemptsTo(
    see.The(question.MetadataOfKey("app/config/database")).IsNot(equal.To(meta)),
)
// or, more specifically:
if updated.(*question.KeyMetadata).Version <= meta.Version {
    log.Fatal("expected the version to be incremented after the update")
}
```

---

## 7. Distributed mutual exclusion (`concurrency/mutex.go`, `concurrency/session.go`)

**Source:** `concurrency.NewSession`, `concurrency.NewMutex(session, prefix)`,
`Mutex.Lock`/`TryLock`/`Unlock`, illustrated by
`tests/v3/integration/clientv3/concurrency/example_mutex_test.go`.

**Why it matters:** The repository's original README already talks about
`app/config/*` keys as if they gate shared configuration; a very natural
next capability for a Screenplay-driven actor is to model "only one actor
may hold the deployment lock at a time" — a textbook etcd use case not
covered at all by the current `ability`/`action` set.

**Proposed contract:**
- New ability `ability.UseDistributedLocking`, built similarly to
  `ability.UseETCD` (a builder `ability.UseLocking().On(session...)`),
  wrapping a `*concurrency.Session` created from the same `*clientv3.Client`
  the actor already uses for `UseETCDAbility`. `Forget` closes the session.
- New actions `action.Lock(name string) *LockAction` and
  `action.Unlock(name string) *UnlockAction`, each resolving a
  `*concurrency.Mutex` for `name` (`concurrency.NewMutex(session, "locks/"+name)`)
  and calling `Lock(ctx)` / `Unlock(ctx)`. `Lock` should return the library's
  own `concurrency.ErrSessionExpired` style errors unwrapped, so tests can
  assert on them directly.

**Example usage:**

```go
alice := screenplay.ActorNamed("alice").WhoCan(
    ability.UseETCD().WithUrls([]string{"localhost:2379"}),
    ability.UseLocking().WithUrls([]string{"localhost:2379"}),
)

err := alice.AttemptsTo(action.Lock("deployment"))
if err != nil {
    log.Fatalf("alice failed to acquire the deployment lock: %v", err)
}

// ... perform the guarded deployment steps ...

err = alice.AttemptsTo(action.Unlock("deployment"))
if err != nil {
    log.Fatalf("alice failed to release the deployment lock: %v", err)
}
```

```gherkin
Scenario: Only one actor can hold the deployment lock at a time
  Given Alice has locked "deployment"
  When Bob attempts to lock "deployment"
  Then Bob fails to acquire the lock
```

---

## 8. Leader election (`concurrency/election.go`)

**Source:** `concurrency.NewElection(session, prefix)`,
`Election.Campaign(ctx, value)`, `Election.Resign(ctx)`,
`Election.Leader(ctx)`, illustrated by
`tests/v3/integration/clientv3/concurrency/example_election_test.go`.

**Why it matters:** A step up from Finding 7: instead of a binary lock,
actors "campaign" to become the leader of a named election and can be
asked who the current leader is — useful for scenarios about singleton
background workers / active-passive failover, a common pattern in
infrastructure automation this extension is meant to support.

**Proposed contract:**
- New actions `action.Campaign(electionName, value string) *CampaignAction`
  (blocks, like `Election.Campaign`, until the actor becomes leader or the
  actor's context is done) and `action.Resign(electionName string) *ResignAction`.
- New question `question.CurrentLeaderOf(electionName string)` wrapping
  `Election.Leader(ctx)`, answering the leader's announced value (`string`)
  or a sentinel "no leader" answer when `concurrency.ErrElectionNoLeader` is
  returned (map this to `(nil, nil)` rather than propagating the "no
  leader" condition as a hard error, since "nobody is leader yet" is a
  valid, assertable state).

**Example usage:**

```go
err := alice.AttemptsTo(action.Campaign("primary-worker", "alice"))
if err != nil {
    log.Fatalf("alice failed to campaign: %v", err)
}

err = alice.AttemptsTo(
    see.The(question.CurrentLeaderOf("primary-worker")).Is(equal.To("alice")),
)

err = alice.AttemptsTo(action.Resign("primary-worker"))
```

---

## 9. Cluster/member health (`maintenance.go`)

**Source:** `Maintenance.Status(ctx, endpoint)`,
`Maintenance.AlarmList(ctx)` in `maintenance.go`, illustrated by
`tests/v3/integration/clientv3/examples/example_maintenance_test.go`.

**Why it matters:** Every scenario in `software/test_cases/core/features/*`
starts with `Given a running ETCD server`; today that precondition is only
verified implicitly (the first `Put`/`Get` would fail). An explicit health
question lets step definitions assert the precondition directly and fail
with a much clearer message ("ETCD is unreachable" vs. a generic Put/Get
error), and lets scenarios about alarms (e.g. "the cluster raises a
NOSPACE alarm") be expressed naturally.

**Proposed contract:**
- New question `question.StatusOfEndpoint(endpoint string)` wrapping
  `Maintenance.Status`, answering the etcd version string (`resp.Version`)
  or similar lightweight summary.
- New question `question.ActiveAlarms()` wrapping `Maintenance.AlarmList`,
  answering `[]clientv3.AlarmMember` (possibly empty, non-nil, mirroring
  the `ValueOfQuestion` empty-collection convention from
  `screenplay_design.md` §6.1).
- Pair the latter with the core library's existing `empty.Collection()`
  resolution (no new resolution needed) to assert "no alarms are active".

**Example usage:**

```go
import "github.com/grandper/go-screenplay/resolution/empty"

err := alice.AttemptsTo(
    see.The(question.StatusOfEndpoint("localhost:2379")).Is(resolution.Exists()),
    see.The(question.ActiveAlarms()).Is(empty.Collection()),
)
if err != nil {
    log.Fatalf("ETCD cluster is not healthy: %v", err)
}
```

```gherkin
Background:
  Given a running ETCD server
  And the ETCD server reports no active alarms
```

---

## 10. Namespacing for test isolation (`namespace/`)

**Source:** `go.etcd.io/etcd/client/v3/namespace` package, referenced from
the client README ("transparently isolate client requests to a
user-defined prefix"), wrapping `KV`, `Lease`, and `Watcher` so every key an
actor touches is automatically prefixed/unprefixed.

**Why it matters:** Every test in `screenplay/action` and
`screenplay/question` currently hand-manages ad-hoc key names
(`"test-key"`, `"app/config/database"`, ...). When these actions are
eventually exercised against a *real*, possibly shared etcd instance (as
hinted at by `go.etcd.io/etcd/tests/v3` and the project's own
`software/test_cases` feature files), parallel test runs risk colliding on
the same keys. Namespacing lets each actor — or each parallel scenario —
operate in its own sandboxed keyspace without changing a single action,
question, or resolution signature.

**Proposed contract:**
- Extend `ability.ETCDBuilder` with `.WithNamespace(prefix string) *ETCDBuilder`,
  which, inside `makeETCDConnection`, wraps the constructed
  `*clientv3.Client`'s `KV`/`Lease`/`Watcher` with
  `namespace.NewKV(client.KV, prefix)` (and the `Lease`/`Watcher`
  equivalents) before storing the result in `RealEtcdClient`. Because
  `ability.EtcdClient` is a narrow interface over `Put`/`Get`/`Delete`/
  `Close`, the namespaced `KV` already satisfies it — no changes needed to
  `action`, `question`, or `resolution`.

**Example usage:**

```go
alice := screenplay.ActorNamed("alice").WhoCan(
    ability.UseETCD().WithNamespace("test-run-42/").WithUrls([]string{"localhost:2379"}),
)

// Alice's actions/questions still refer to plain keys...
err := alice.AttemptsTo(action.PutTheValue("v1").ForTheKey("app/config/database"))
// ...but on the wire etcd actually stores/reads "test-run-42/app/config/database",
// fully isolated from any other concurrently running scenario.
```

---

## 11. Gaps discovered from `pkg/store/storeetcd` itself

The earlier sections are useful, but they do not yet cover the full API surface that `pkg/store/storeetcd/storeetcd.go` actually exposes. A screenplay implementation that aims to test this package end-to-end needs to model the following capabilities as first-class language, not only the simplest Put/Get/Delete flow.

### 11.1 Batch and prefix operations are required, not optional

The real store implementation adds several operations that are immediately exercised by `storeetcd_test.go` and are not covered well by the earlier design:

- `Gets(ctx, keys []string)` and `GetsWithRevisions(ctx, keys []string)`
- `Puts(ctx, values map[string][]byte)` and `PutsWithRevisions(ctx, values, revisions)`
- `Deletes(ctx, keys []string)` and `DeletesWithRevisions(ctx, keys, revisions)`
- `ListKeysWithPrefix(ctx, prefix string)`
- `DeleteKeysWithPrefix(ctx, prefix string)`
- `DeleteKeysByCriteria(ctx, prefix string, criteria store.Criteria)`

These are not minor wrappers around a single-key API; they exercise the transaction-heavy parts of etcd and are explicitly tested for:

- max transaction size (`ErrTooManyOperationsOnTx`)
- empty mixed input sets
- partial misses in multi-key reads
- prefix delete/list behaviour
- criteria-based deletion of a subset of keys under a prefix

A screenplay extension that only supports `PutTheValue` / `ValueForKey` will not be able to reproduce these tests faithfully.

### 11.2 The design needs explicit coverage for revision-aware compare-and-swap contracts

`storeetcd.go` is built around etcd revision semantics. In particular:

- `GetsWithRevisions` returns both values and their `ModRevision`
- `PutsWithRevisions` performs a conditional transaction and returns the new revisions
- `DeletesWithRevisions` performs conditional delete if prior revisions still match
- `DeleteKeysByCriteria` is a read-then-delete flow over a prefix that must be batched

The current enrichment document sketches conditional writes in a general way, but it does not make the following requirement explicit:

- the screenplay library must be able to answer "what is the current revision of this key?"
- the screenplay library must be able to perform both "conditional put" and "conditional delete" operations from the actor point of view
- tests must be able to assert `store.ErrResourceModified`-style failures (`!response.Succeeded`), not only raw `Put`/`Delete` errors

This matters because the package's own tests actively verify optimistic concurrency through revision comparisons.

### 11.3 TTL expiry, watches, and lease-aware deletion are a mandatory test surface

The existing design includes TTL and watch concepts, but the actual package tests go further and combine them:

- `Put` supports a default TTL via a lease
- `WatchEvents` observes prefix-based changes and preserves `PrevKv`
- `IsEventExpired` distinguishes lease expiry from a normal delete
- tests assert that a key deleted because its lease expired is detected as an expiration event

This means the screenplay layer must support not just a simple "put with TTL" action, but the ability to:

- watch a prefix for change events
- inspect event payloads (`Event.Type`, `Event.Kv`, `Event.PrevKv`)
- tell whether a delete was caused by lease expiry
- assert on the absence or presence of that condition

Without this, the TTL expiry tests in `TestIsEventExpired` cannot be re-created using screenplay semantics.

### 11.4 The `screenplay` layer must support "prefix + value query" semantics, not just a single key

The store implementation has operations to enumerate keys and to delete or filter by prefix, and tests explicitly verify these semantics:

- `ListKeysWithPrefix` returns every key under a prefix
- `DeleteKeysWithPrefix` removes everything under a prefix
- `DeleteKeysByCriteria` filters the subset under a prefix using a value matcher

The enrichment document mentions `question.NumberOfKeysMatching(prefix)` and `ValuesStartingWithPrefix`, but it still lacks the direct actor vocabulary for:

- listing matching keys under a prefix
- deleting all keys under a prefix
- deleting only keys under a prefix that satisfy a predicate on the value

This is a concrete gap relative to the package's tests.

### 11.5 The library must support a readiness probe, not only connectivity and CRUD

`StoreEtcd.IsReady(ctx, timeout)` is a first-class store operation and is used in the test suite:

- `TestIsReady` expects success against a live cluster
- `TestNewStore` expects connection errors for empty/malformed URLs and canceled contexts

The design may model cluster health at the etcd layer, but for this package to be testable as a true store abstraction, the screenplay library should also expose a readiness question or precondition around "the store is ready to accept operations". This keeps tests from depending on a bare first `Put` as the only readiness signal.

### 11.6 Required Screenplay vocabulary to test `storeetcd` completely

To cover the existing `storeetcd` API surface, the screenplay extension should include at least the following additional vocabulary beyond the core PUT/GET/DELETE proposals:

- `action.PutManyValues(values map[string][]byte)`
- `action.PutManyValuesIfUnchanged(values, revisions)`
- `action.DeleteKeys(keys []string)`
- `action.DeleteKeysIfUnchanged(keys, revisions)`
- `action.DeleteKeysWithPrefix(prefix string)`
- `action.DeleteKeysByCriteria(prefix string, criteria store.Criteria)`
- `question.KeysWithPrefix(prefix string)`
- `question.ValuesForKeys(keys []string)`
- `question.ValuesWithRevisions(keys []string)`
- `question.MetadataOfKey(key string)` or `MetadataForKeys(keys []string)`
- `action.WatchPrefix(prefix string)` / `action.WaitUntilPrefixChanges(prefix string)`
- `question.LatestEventOnPrefix(prefix string)`
- `resolution.WasLeaseExpired()`
- `question.IsStoreReady(timeout time.Duration)`

This is not a wish list from a broader etcd product surface; it is a direct reflection of methods that already exist and are tested in `pkg/store/storeetcd`.

### 11.7 Testability requirement for the final implementation

The design should be considered complete only if these test scenarios can be expressed using screenplay actors, without reaching into the concrete `*storeetcd.StoreEtcd` type directly:

- put and get a single value
- batch put/get several values in one transaction
- read multiple keys and fetch their revisions
- write multiple keys only if each key still has the expected revision
- delete a single key and multiple keys
- delete under a prefix and delete selected entries by criteria
- list keys under a prefix
- watch a prefix and assert a put or delete event occurred
- detect a delete caused by a lease expiry
- assert readiness of the store and failure on invalid config
- ensure the transaction limit is enforced and invalid empty keys are rejected

If the Screenplay abstraction cannot express one of these scenarios, the design remains incomplete for this repository's actual etcd store contract.

### 11.8 Summary

The core missing theme is this: the original screenplay design is strongly oriented around one-key CRUD. `pkg/store/storeetcd` is not just a one-key store; it is a transaction-heavy, prefix-aware, revision-aware, watch-aware, lease-aware store. Any screenplay extension intended to support this package must include the stronger multi-key, batch, conditional-write/delete, and event-detection vocabulary above, otherwise the package's own tests cannot be reproduced in a clear, actor-driven way.
