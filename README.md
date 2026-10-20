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
| `action` | `github.com/grandper/go-screenplay-etcd/action` | `ErrKeyNotProvided`, `ErrPrefixNotProvided`, `ErrKeyAlreadyExists`, `ErrKeyModified` |
| `put` | `github.com/grandper/go-screenplay-etcd/action/put` | `TheValue`, `TheValues` |
| `remove` | `github.com/grandper/go-screenplay-etcd/action/remove` | `TheKey`, `TheKeys`, `TheKeysStartingWith` |
| `question` | `github.com/grandper/go-screenplay-etcd/question` | `KeyNamed`, `ValueOfTheKey`, `RevisionOfTheKey`, `TimeToLiveOfTheKey`, `KeysStartingWith`, `ValuesOfTheKeysStartingWith`, `NumberOfKeysStartingWith`, `ErrKeyNotFound`, `ErrKeyNotProvided`, `ErrPrefixNotProvided` |
| `fixture` (questions) | `github.com/grandper/go-screenplay-etcd/question/fixture` | `CreateFakeExistingKey`, `CreateFakeMissingKey` |
| `resolution` | `github.com/grandper/go-screenplay-etcd/resolution` | `Exist` |
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

### Putting a value that expires
A key can be given a time to live: the value is put with a lease, and the store removes
the key when the lease expires. Close the statement with `ExpiringAfter` and a unit
(`Seconds`, `Minutes`, or `Milliseconds`, with their singular forms); etcd counts in whole
seconds, so the time is rounded up:
```go
err := anActor.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(10).Seconds())
// Alice attempts to put the value "held" for the key "locks/batch-job" expiring after 10 seconds
```
A long-running actor, such as a worker holding a soft lock, renews the lease instead of
letting it expire. `AndKeepItAlive` renews it in the background for as long as the actor
keeps the ability, so the key disappears only once the actor exits:
```go
err := anActor.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(10).Seconds().AndKeepItAlive())
// Alice attempts to put the value "held" for the key "locks/batch-job" expiring after 10 seconds and keep it alive

err = anActor.Exit() // the renewals stop, and the key expires ten seconds later
```
The expiry combines with the conditions below: a lock is acquired by putting a value only
if the key is absent, expiring after a while.

### Putting several values at once
Put several values, each for its key, with `put.TheValues`. They are put in one
transaction, so the store never holds some of them without the others, and the narration
lists them sorted by key:
```go
err := anActor.AttemptsTo(put.TheValues(map[string]string{
    "app/config/database": "postgresql://localhost:5432",
    "app/config/cache":    "redis://localhost:6379",
}))
// Alice attempts to put the values "app/config/cache" = "redis://localhost:6379", "app/config/database" = "postgresql://localhost:5432"
```
An empty map, or a map holding an empty key, is reported with `ErrKeyNotProvided`. The
values can be kept secret the same way as a single one; the keys are still narrated:
```go
err := anActor.AttemptsTo(put.TheValues(credentials).Secretly())
// Alice attempts to put the values <secret> for the keys "app/config/password", "app/config/username"
```

### Writing conditionally
`put.TheValue` always overwrites the value a key holds. To initialize a key only when no
one else did, close the statement with `IfTheKeyIsAbsent`: the check and the put happen
in one transaction, and a key that exists makes the action fail with an error wrapping
`ErrKeyAlreadyExists`, leaving the store untouched:
```go
err := anActor.AttemptsTo(put.TheValue("initial").ForTheKey("app/config/flag").IfTheKeyIsAbsent())
// Alice attempts to put the value "initial" for the key "app/config/flag" if the key is absent
// fails: the key "app/config/flag" already exists
```
To update a key only when nobody changed it since you read it, ask for its revision with
`RevisionOfTheKey` (see below), then close the statement with
`IfTheKeyIsUnchangedSinceTheRevision`. A key modified in between makes the action fail
with an error wrapping `ErrKeyModified`, so a read-modify-write loop knows to read again:
```go
revision, _ := anActor.AsksFor(RevisionOfTheKey("app/config/counter"))

err := anActor.AttemptsTo(put.TheValue("2").ForTheKey("app/config/counter").IfTheKeyIsUnchangedSinceTheRevision(revision))
// Alice attempts to put the value "2" for the key "app/config/counter" if the key is unchanged since the revision 42
// fails: the key "app/config/counter" was modified since the revision 42
```
The revision zero means that the key must not exist, which is another way of saying
`IfTheKeyIsAbsent`. The same condition closes `remove.TheKey`, with
`IfUnchangedSinceTheRevision`:
```go
err := anActor.AttemptsTo(remove.TheKey("app/config/counter").IfUnchangedSinceTheRevision(revision))
// Alice attempts to remove the key "app/config/counter" if unchanged since the revision 42
```

### Removing keys
Remove a key with `remove.TheKey`. Removing a key that does not exist is not an error, so
a cleanup task never fails for a key a scenario did not create:
```go
err := anActor.AttemptsTo(remove.TheKey("app/config/cache"))
// Alice attempts to remove the key "app/config/cache"
```
Remove several keys at once with `remove.TheKeys`. They are removed in one transaction,
so the store never holds some of them without the others:
```go
err := anActor.AttemptsTo(remove.TheKeys("app/config/cache", "app/config/database"))
// Alice attempts to remove the keys "app/config/cache", "app/config/database"
```
Remove every key starting with a prefix with `remove.TheKeysStartingWith`:
```go
err := anActor.AttemptsTo(remove.TheKeysStartingWith("app/config/"))
// Alice attempts to remove the keys starting with "app/config/"
```
An empty key, or no key at all, is reported when the actor performs the action, with
`ErrKeyNotProvided`. An empty prefix, which would reach every key of the store, is reported
with `ErrPrefixNotProvided`:
```go
err := anActor.AttemptsTo(remove.TheKeys())
// the key must be provided
err := anActor.AttemptsTo(remove.TheKeysStartingWith(""))
// the prefix must be provided
```

## Questions

### The value of a key
`ValueOfTheKey` answers with the value the key holds, a `string`, so you assert on it like
any text:
```go
err := anActor.Should(see.The(ValueOfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
// Alice should see that the value of the key "app/config/database" should be equal to "postgresql://localhost:5432"
```
A key that does not exist makes the question fail, with an error wrapping `ErrKeyNotFound`,
and an empty key with `ErrKeyNotProvided`:
```go
value, err := anActor.AsksFor(ValueOfTheKey("does/not/exist"))
// the key "does/not/exist" was not found
```

### A key and its metadata
`KeyNamed` answers with an `*ability.Key`, whether the key exists or not, so you assert on
its existence with the `Exist` resolution of the extension:
```go
err := anActor.Should(see.The(KeyNamed("app/status")).Should(Exist()))
// Alice should see that the key named "app/status" should exist
err := anActor.Should(see.The(KeyNamed("app/status")).ShouldNot(Exist()))
// Alice should see that the key named "app/status" should not exist
// fails: Alice failed to see that the key named "app/status" should exist: it does not exist
```
An `*ability.Key` exposes its parts through getters: `Name()` and `Exists()`, then
`Value()`, `Version()` (the number of times the key was put since it was created),
`CreateRevision()` and `ModRevision()` (the revisions of the store at which it was created
and last modified), and `Lease()` (the ID of the lease attached to it, zero without a
lease). They are all zero or empty when the key does not exist:
```go
key, _ := anActor.AsksFor(KeyNamed("app/config/database")) // key is an *ability.Key

if key.Exists() && key.Version() > 1 {
    fmt.Printf("%s was updated %d times, last at revision %d\n", key.Name(), key.Version()-1, key.ModRevision())
}
```
`RevisionOfTheKey` answers directly with the revision at which the key was last modified,
an `int64`, which is what a conditional write needs. It fails like `ValueOfTheKey` when
the key does not exist:
```go
revision, err := anActor.AsksFor(RevisionOfTheKey("app/config/database"))
```
`TimeToLiveOfTheKey` answers with the time the key has left before its lease expires and
removes it, a `time.Duration` in whole seconds. A key without a lease never expires, and
has zero time to live:
```go
err := anActor.Should(see.The(TimeToLiveOfTheKey("locks/batch-job")).Should(be.AtMost(10*time.Second)))
// Alice should see that the time to live of the key "locks/batch-job" should be at most 10s
err := anActor.Should(see.The(TimeToLiveOfTheKey("app/config/database")).Should(be.Zero[time.Duration]()))
```

### The keys starting with a prefix
Three questions look at the keys starting with a prefix. `KeysStartingWith` answers with
the keys, a `[]string` in the order of the keys; `ValuesOfTheKeysStartingWith` answers
with a `map[string]string` from each key to its value; and `NumberOfKeysStartingWith`
answers with how many there are, an `int`, counted by the store without retrieving them:
```go
err := anActor.Should(see.The(KeysStartingWith("app/config/")).Should(contain.Element("app/config/database")))
// Alice should see that the keys starting with "app/config/" should contain the element "app/config/database"
err := anActor.Should(see.The(ValuesOfTheKeysStartingWith("app/config/")).Should(contain.Key[string, string]("app/config/database")))
err := anActor.Should(see.The(NumberOfKeysStartingWith("app/config/")).Should(be.EqualTo(2)))
// Alice should see that the number of keys starting with "app/config/" should be equal to 2
```
When no key starts with the prefix, the slice and the map are empty, and the number is
zero: none of the questions fails.
```go
err := anActor.Should(see.The(ValuesOfTheKeysStartingWith("app/config/")).Should(be.Empty[map[string]string]()))
```
An empty prefix is reported with `ErrPrefixNotProvided`.

## Resolutions

### Existence
`Exist` tells if the key answered by `KeyNamed` exists. It is a bare infinitive and takes
the verb `Should` or `ShouldNot`; a nil key does not exist:
```go
err := anActor.Should(see.The(KeyNamed("app/status")).Should(Exist()))
// Alice should see that the key named "app/status" should exist
```
To test a resolution or an action of your own against a key without a store, the
`question/fixture` package creates fake keys:
```go
existing := fixture.CreateFakeExistingKey("app/status", "ready")
missing := fixture.CreateFakeMissingKey("does/not/exist")
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
