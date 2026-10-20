# go-screenplay etcd extension

This extension of [go-screenplay](https://github.com/grandper/go-screenplay) lets an
actor drive an [etcd](https://etcd.io) key-value store. It bundles the ability to use
etcd together with the actions, questions, and resolutions that go with it, so you test
a system built on etcd with the same screenplay vocabulary you use everywhere else.

## Installation
```bash
go get github.com/grandper/go-screenplay-etcd
```
The extension requires Go 1.27 or later, and drives etcd v3 through its official client,
`go.etcd.io/etcd/client/v3`.

## Packages
The symbols shown below live in the extension's sub-packages:

| Package | Import path | Provides |
|---------|-------------|----------|
| `ability` | `github.com/grandper/go-screenplay-etcd/ability` | `UseETCD`, `Key`, `ErrNoEndpointProvided`, `ErrFailedToConnect` |
| `action` | `github.com/grandper/go-screenplay-etcd/action` | `ErrKeyNotProvided` |
| `put` | `github.com/grandper/go-screenplay-etcd/action/put` | `TheValue` |
| `question` | `github.com/grandper/go-screenplay-etcd/question` | `ValueOfTheKey`, `ErrKeyNotFound` |
| `fixture` | `github.com/grandper/go-screenplay-etcd/fixture` | `RunETCDServer`, `ETCDServer` |

Assertions are written with the core screenplay packages (`action/see`,
`resolution/be`, `resolution/contain`, `resolution/length`, ...). The actions are written
with their package, which is their verb: `put.TheValue(...)` reads "put the value". For
readability the examples below drop the package qualifiers on the other extension
constructors; import the packages above and qualify them as `ability.UseETCD()`, and so on.
The extension's `question` package shares its name with the core one: when a file uses
both, keep the core name and alias the extension package with the `etcd` prefix:
```go
import (
	"github.com/grandper/go-screenplay/v2/action/keep"
	"github.com/grandper/go-screenplay-etcd/action/put"
)

err := anActor.AttemptsTo(keep.TryingTo(put.TheValue("ready").ForTheKey("app/status")))
```

## The ability
This extension introduces a single ability: using etcd. Give it to an actor with
`WhoCan`, together with the endpoints of the cluster, in the form `host:port`:
```go
anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheEndpoints("localhost:2379"))
anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheEndpoints("etcd-1:2379", "etcd-2:2379", "etcd-3:2379"))
```
The connection is opened the first time the actor uses the ability, so an ability given
no endpoint is reported by the first action, with `ErrNoEndpointProvided`, and a client
that cannot be created with an error wrapping `ErrFailedToConnect`:
```go
anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD())

err := anActor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// no etcd endpoint was provided to the ability
```

The ability is configured fluently, and the chain reads as English: "use etcd with the
endpoints localhost:2379 with the namespace test-run-42/ with timing out after 5 seconds":
```go
anActor := screenplay.ActorNamed("Alice").WhoCan(
    UseETCD().
        WithTheEndpoints("localhost:2379").
        WithTheNamespace("test-run-42/").
        WithTimingOutAfter(5).Seconds(),
)
```
`WithTheNamespace` prefixes every key the actor uses. The actions and the questions keep
using plain keys: the prefix is added on the wire and removed from what the store answers,
so concurrent scenarios run against one cluster without sharing keys:
```go
err := anActor.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"))
// the store holds the key "test-run-42/app/config/database"
```
`WithTimingOutAfter` bounds every operation of the actor. Without it, an operation is
bounded only by the deadline of the actor's context (see `WithContext`). An operation that
outlasts the timeout fails, with `context.DeadlineExceeded`:
```go
err := anActor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// context deadline exceeded
```
The timeout is closed by its unit: `Milliseconds`, `Seconds` or `Minutes`, each with
its singular form (`Millisecond`, `Second`, `Minute`) so that one unit reads as English:
```go
anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(500).Milliseconds())
anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(1).Minute())
```
A client the test created, for example with authentication or with options the ability
does not expose, is given with `WithTheClient` instead of the endpoints. The namespace, if
any, is set on that client, and the ability owns it from then on:
```go
client, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}, Username: "alice", Password: "s3cret"})

anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheClient(client))
```

The settings are the only methods a test can call on the ability. Its operations live in
`internal/ability`, so only an action or a question of the extension uses them, in one
statement written in `PerformAs` or `AnsweredBy` that reads "the ability of the actor to
put":
```go
err := impl.TheAbilityOf(theActor).ToPut(theActor.Context(), "app/config/database", "postgresql://localhost:5432")
```
`TheAbilityOf` never fails: when the actor cannot use etcd, the operation reports
`screenplay.ErrActorMissingAbility`.

When the actor exits, the ability forgets the session: the connection to the store is
closed, whether it was opened by the ability or given with `WithTheClient`.
```go
err := anActor.Exit()
```

## Actions

### Putting a value
Put a value for a key with `put.TheValue`, closing the statement with `ForTheKey`. The key
is created when it does not exist, and its value is overwritten when it does:
```go
err := anActor.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"))
// Alice attempts to put the value "postgresql://localhost:5432" for the key "app/config/database"
```
Nothing panics when the statement is written wrongly. A missing or empty key is reported
when the actor performs the action, with `ErrKeyNotProvided`:
```go
err := anActor.AttemptsTo(put.TheValue("postgresql://localhost:5432"))
// the key must be provided
```
If the value is a secret you do not want in the narration, mark the action secret with
`Secretly` (or its longer alias `WhichShouldBeKeptSecret`):
```go
err := anActor.AttemptsTo(put.TheValue(password).ForTheKey("app/config/password").Secretly())
err := anActor.AttemptsTo(put.TheValue(password).ForTheKey("app/config/password").WhichShouldBeKeptSecret())
// Alice attempts to put the value <secret> for the key "app/config/password"
```

## Questions

### The value of a key
`ValueOfTheKey` answers with the value the key holds, a `string`, so you assert on it like
any text:
```go
err := anActor.Should(see.The(ValueOfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
// Alice should see that the value of the key "app/config/database" should be equal to "postgresql://localhost:5432"
```
A key that does not exist makes the question fail, with an error wrapping `ErrKeyNotFound`:
```go
value, err := anActor.AsksFor(ValueOfTheKey("does/not/exist"))
// the key "does/not/exist" was not found
```

## Testing with an embedded etcd server
The `fixture` package spins up an in-process etcd server, so scenarios run without an
external cluster and without Docker. `RunETCDServer` takes a `*testing.T` and registers
its own cleanup, so the server is stopped when the test ends. Every call starts its own
server on a free port of the loopback interface, so tests never share keys:
```go
server := fixture.RunETCDServer(t)
anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheEndpoints(server.Endpoint()))
```
The server keeps its data in the temporary directory of the test and writes its logs to
a file there (`server.LogFile()`), so they never mix with the output of the test.

`NewClient` returns a client connected to the server, closed when the test ends, to check
the state of the store from outside the actor:
```go
client := server.NewClient(t)

response, err := client.Get(ctx, "app/config/database")
```

## License
This extension is released under the [MIT License](LICENSE).
