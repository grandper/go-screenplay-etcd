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
| `ability` | `github.com/grandper/go-screenplay-etcd/ability` | `UseETCD`, `Key`, `Event`, `EndpointStatus`, `Alarm`, `ErrNoEndpointProvided`, `ErrFailedToConnect`, `ErrFailedToLoadTheCertificates`, `ErrAlreadyWatching`, `ErrNotWatching`, `ErrLocked`, `ErrLockAlreadyHeld`, `ErrLockNotHeld`, `ErrAlreadyCampaigning`, `ErrNotCampaigning` |
| `action` | `github.com/grandper/go-screenplay-etcd/action` | `ErrKeyNotProvided`, `ErrPrefixNotProvided`, `ErrNameNotProvided`, `ErrKeyAlreadyExists`, `ErrKeyModified` |
| `put` | `github.com/grandper/go-screenplay-etcd/action/put` | `TheValue`, `TheValues` |
| `remove` | `github.com/grandper/go-screenplay-etcd/action/remove` | `TheKey`, `TheKeys`, `TheKeysStartingWith` |
| `watch` | `github.com/grandper/go-screenplay-etcd/action/watch` | `TheKey`, `TheKeysStartingWith` |
| `acquire` | `github.com/grandper/go-screenplay-etcd/action/acquire` | `TheLock` |
| `release` | `github.com/grandper/go-screenplay-etcd/action/release` | `TheLock` |
| `campaign` | `github.com/grandper/go-screenplay-etcd/action/campaign` | `InTheElection` |
| `resign` | `github.com/grandper/go-screenplay-etcd/action/resign` | `FromTheElection` |
| `question` | `github.com/grandper/go-screenplay-etcd/question` | `ErrKeyNotFound`, `ErrKeyNotProvided`, `ErrPrefixNotProvided`, `ErrNameNotProvided`, `ErrEndpointNotProvided` |
| `value` | `github.com/grandper/go-screenplay-etcd/question/value` | `OfTheKey` |
| `key` | `github.com/grandper/go-screenplay-etcd/question/key` | `Named` |
| `revision` | `github.com/grandper/go-screenplay-etcd/question/revision` | `OfTheKey` |
| `ttl` | `github.com/grandper/go-screenplay-etcd/question/ttl` | `OfTheKey` |
| `keys` | `github.com/grandper/go-screenplay-etcd/question/keys` | `StartingWith` |
| `values` | `github.com/grandper/go-screenplay-etcd/question/values` | `OfTheKeysStartingWith` |
| `events` | `github.com/grandper/go-screenplay-etcd/question/events` | `OnTheKey`, `OnTheKeysStartingWith` |
| `leader` | `github.com/grandper/go-screenplay-etcd/question/leader` | `OfTheElection` |
| `status` | `github.com/grandper/go-screenplay-etcd/question/status` | `OfTheEndpoint` |
| `alarms` | `github.com/grandper/go-screenplay-etcd/question/alarms` | `OfTheCluster` |
| `fixture` (questions) | `github.com/grandper/go-screenplay-etcd/question/fixture` | `CreateFakeExistingKey`, `CreateFakeMissingKey`, `CreateFakePutEvent`, `CreateFakeDeleteEvent`, `CreateFakeExpiryEvent`, `CreateFakeAlarm` |
| `exist` | `github.com/grandper/go-screenplay-etcd/resolution/exist` | `InETCD` |
| `been` | `github.com/grandper/go-screenplay-etcd/resolution/been` | `Put`, `Deleted` |
| `fixture` | `github.com/grandper/go-screenplay-etcd/fixture` | `RunETCDServer`, `RunSecureETCDServer`, `ETCDServer` |

Assertions are written with the core screenplay packages (`action/see`,
`resolution/be`, `resolution/contain`, `resolution/length`, ...). Every statement is
written with its packages, which are words of the sentence: the package of an action is
its verb, `put.TheValue("ready").ForTheKey("app/status")` reads "put the value ready for
the key app/status"; the package of a question is its first word,
`value.OfTheKey("app/status")` reads "value of the key app/status"; and the package of a
resolution is the word that follows the verb, `exist.InETCD()` reads "exist in etcd"
after `Should`. Read aloud, a whole statement is one English sentence, which is what the
actor narrates:
```go
err := anActor.Should(see.The(value.OfTheKey("app/status")).Should(be.EqualTo("ready")))
// Alice should see that the value of the key "app/status" should be equal to "ready"
```
The examples below are written as you write them, package qualifiers included; only
`ability.UseETCD()` is shortened to `UseETCD()`. The extension's `fixture` packages
share their name with core ones: when a file uses both, keep the core name and alias the
extension package with the `etcd` prefix:
```go
import (
	"github.com/grandper/go-screenplay/v2/question/last"
	"github.com/grandper/go-screenplay/v2/resolution/fixture"
	etcdfixture "github.com/grandper/go-screenplay-etcd/question/fixture"
	"github.com/grandper/go-screenplay-etcd/resolution/been"
)

fixture.AssertMatch(t, been.Put(), etcdfixture.CreateFakePutEvent("app/status", "ready"))
```
The core actions compose with the ones of the extension as with any other: `keep.TryingTo`
retries a put against a cluster that is still starting, `proceed.ConcurrentlyTo` runs
several of them at once, and `wait.UntilThe` waits for a question of the extension to
answer as expected:
```go
err := anActor.AttemptsTo(keep.TryingTo(put.TheValue("ready").ForTheKey("app/status")).For(30).Seconds())
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
A cluster that secures its connections with TLS is given the authority that signed its
certificates with `WithTheCertificateAuthority`; without it, the authorities of the system
are trusted. A cluster that authenticates its clients with mutual TLS is also given the
certificate the actor presents, with `WithTheClientCertificate` closed by `AndTheKey`. All
three are PEM files:
```go
anActor := screenplay.ActorNamed("Alice").WhoCan(
    UseETCD().
        WithTheEndpoints("etcd.example.com:2379").
        WithTheCertificateAuthority("ca.crt").
        WithTheClientCertificate("client.crt").AndTheKey("client.key"),
)
```
A file that cannot be read, or that holds no certificate, is reported by the first action,
with an error wrapping both `ErrFailedToConnect` and `ErrFailedToLoadTheCertificates`:
```go
err := anActor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// failed to connect to the etcd server: failed to load the certificates: open client.crt: no such file or directory
```
A client the test created, for example with authentication or with options the ability
does not expose, is given with `WithTheClient` instead of the endpoints. The namespace, if
any, is set on that client, and the ability owns it from then on:
```go
client, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}, Username: "alice", Password: "s3cret"})

anActor := screenplay.ActorNamed("Alice").WhoCan(UseETCD().WithTheClient(client))
```

When the actor exits, the ability also stops the watches it started and the renewals of
the leases it keeps alive, releases the locks it holds and resigns from the elections it
campaigns in.

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
`revision.OfTheKey` (see below), then close the statement with
`IfTheKeyIsUnchangedSinceTheRevision`. A key modified in between makes the action fail
with an error wrapping `ErrKeyModified`, so a read-modify-write loop knows to read again:
```go
revision, _ := anActor.AsksFor(revision.OfTheKey("app/config/counter"))

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

### Watching keys
Many scenarios are event-driven: another process updates a configuration key, and the
actor should notice. Rather than polling, watch the key with `watch.TheKey`: every change
made to it from then on is recorded, until the actor exits. The action returns once the
store confirmed the watch, so a change made right after it is never missed:
```go
err := anActor.AttemptsTo(watch.TheKey("app/config/database"))
// Alice attempts to watch the key "app/config/database"
```
Watch every key starting with a prefix with `watch.TheKeysStartingWith`:
```go
err := anActor.AttemptsTo(watch.TheKeysStartingWith("app/config/"))
// Alice attempts to watch the keys starting with "app/config/"
```
A key, or a prefix, the actor already watches is reported with an error wrapping
`ErrAlreadyWatching`; a watch of a key and a watch of the keys starting with it are two
different watches:
```go
err := anActor.AttemptsTo(watch.TheKey("app/config/database"), watch.TheKey("app/config/database"))
// the key "app/config/database" is already watched
```
The recorded changes are asked about with `events.OnTheKey` and
`events.OnTheKeysStartingWith` (see below), and waited for with the core `wait.UntilThe`.

### Locking
Only one actor at a time may hold a distributed lock. `acquire.TheLock` acquires the lock
of the given name, waiting for the actor holding it to release it; the wait ends with the
context of the actor (see `WithContext`). The lock is released with `release.TheLock`, or
when the actor exits:
```go
err := anActor.AttemptsTo(acquire.TheLock("deployment"))
// Alice attempts to acquire the lock "deployment"
err = anActor.AttemptsTo(release.TheLock("deployment"))
// Alice attempts to release the lock "deployment"
```
To fail instead of waiting when another actor holds the lock, close the statement with
`WithoutWaiting`; the action then fails with `ErrLocked`, the error of the etcd client:
```go
err := anotherActor.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting())
// Bob attempts to acquire the lock "deployment" without waiting
// fails: mutex: Locked by another session
```
Acquiring a lock the actor already holds is reported with an error wrapping
`ErrLockAlreadyHeld`, and releasing a lock it does not hold with an error wrapping
`ErrLockNotHeld`; an empty name with `ErrNameNotProvided`:
```go
err := anActor.AttemptsTo(release.TheLock("deployment"))
// the lock "deployment" is not held by the actor
```
The locks of an actor share one session, a lease the client keeps alive: should the actor
lose its connection to the cluster for longer than the session allows, the store releases
its locks.

### Campaigning in an election
A step up from a lock, an election names a leader among the actors campaigning in it, and
anyone can ask who leads. `campaign.InTheElection` campaigns in the election of the given
name and returns once the actor leads it, waiting for the leader to resign if there is
one. The actor is announced under its name, unless it campaigns `WithTheValue` of its own.
The leadership is given up with `resign.FromTheElection`, or when the actor exits:
```go
err := anActor.AttemptsTo(campaign.InTheElection("primary-worker"))
// Alice attempts to campaign in the election "primary-worker"
err := anActor.AttemptsTo(campaign.InTheElection("primary-worker").WithTheValue("node-1"))
// Alice attempts to campaign in the election "primary-worker" with the value "node-1"
err = anActor.AttemptsTo(resign.FromTheElection("primary-worker"))
// Alice attempts to resign from the election "primary-worker"
```
Campaigning in an election the actor already campaigns in is reported with an error
wrapping `ErrAlreadyCampaigning`, and resigning from an election it does not campaign in
with an error wrapping `ErrNotCampaigning`:
```go
err := anActor.AttemptsTo(resign.FromTheElection("primary-worker"))
// the actor does not campaign in the election "primary-worker"
```

## Questions

### The value of a key
`value.OfTheKey` answers with the value the key holds, a `string`, so you assert on it
like any text:
```go
err := anActor.Should(see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
// Alice should see that the value of the key "app/config/database" should be equal to "postgresql://localhost:5432"
```
A key that does not exist makes the question fail, with an error wrapping `ErrKeyNotFound`,
and an empty key with `ErrKeyNotProvided`:
```go
missing, err := anActor.AsksFor(value.OfTheKey("does/not/exist"))
// the key "does/not/exist" was not found
```

### A key and its metadata
`key.Named` answers with an `*ability.Key`, whether the key exists or not, so you assert
on its existence with the `exist.InETCD` resolution of the extension:
```go
err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
// Alice should see that the key named "app/status" should exist in etcd
err := anActor.Should(see.The(key.Named("app/status")).ShouldNot(exist.InETCD()))
// Alice should see that the key named "app/status" should not exist in etcd
// fails: Alice failed to see that the key named "app/status" should exist in etcd: it does not exist
```
An `*ability.Key` exposes its parts through getters: `Name()` and `Exists()`, then
`Value()`, `Version()` (the number of times the key was put since it was created),
`CreateRevision()` and `ModRevision()` (the revisions of the store at which it was created
and last modified), and `Lease()` (the ID of the lease attached to it, zero without a
lease). They are all zero or empty when the key does not exist:
```go
database, _ := anActor.AsksFor(key.Named("app/config/database")) // database is an *ability.Key

if database.Exists() && database.Version() > 1 {
    fmt.Printf("%s was updated %d times, last at revision %d\n", database.Name(), database.Version()-1, database.ModRevision())
}
```
`revision.OfTheKey` answers directly with the revision at which the key was last
modified, an `int64`, which is what a conditional write needs. It fails like
`value.OfTheKey` when the key does not exist:
```go
lastModification, err := anActor.AsksFor(revision.OfTheKey("app/config/database"))
// Alice asks for the revision of the key "app/config/database"
```
`ttl.OfTheKey` answers with the time to live of the key, the time it has left before its
lease expires and removes it, a `time.Duration` in whole seconds. A key without a lease
never expires, and has zero time to live:
```go
err := anActor.Should(see.The(ttl.OfTheKey("locks/batch-job")).Should(be.AtMost(10*time.Second)))
// Alice should see that the TTL of the key "locks/batch-job" should be at most 10s
err := anActor.Should(see.The(ttl.OfTheKey("app/config/database")).Should(be.Zero[time.Duration]()))
```

### The keys starting with a prefix
Two questions look at the keys starting with a prefix. `keys.StartingWith` answers with
the keys, a `[]string` in the order of the keys, and `values.OfTheKeysStartingWith`
answers with a `map[string]string` from each key to its value. Count the keys with the
core `number.Of`:
```go
err := anActor.Should(see.The(keys.StartingWith("app/config/")).Should(contain.Element("app/config/database")))
// Alice should see that the keys starting with "app/config/" should contain the element "app/config/database"
err := anActor.Should(see.The(values.OfTheKeysStartingWith("app/config/")).Should(contain.Key[string, string]("app/config/database")))
// Alice should see that the values of the keys starting with "app/config/" should contain the key "app/config/database"
err := anActor.Should(see.The(number.Of(keys.StartingWith("app/config/"))).Should(be.EqualTo(2)))
// Alice should see that the number of keys starting with "app/config/" should be equal to 2
```
When no key starts with the prefix, the slice and the map are empty: neither question
fails.
```go
err := anActor.Should(see.The(values.OfTheKeysStartingWith("app/config/")).Should(be.Empty[map[string]string]()))
```
An empty prefix is reported with `ErrPrefixNotProvided`.

### The events on a watched key
`events.OnTheKey` answers with the changes recorded on a key the actor watches, as a
`[]*ability.Event` in the order they happened; `events.OnTheKeysStartingWith` does the
same for a watched prefix. Both answer with an empty slice when nothing happened yet, and
fail with an error wrapping `ErrNotWatching` when the actor does not watch the key or the
prefix. Because they answer with a slice, they compose with the core `wait.UntilThe`,
`length`, `first` and `last` packages, and with the `been` resolutions of the extension:
```go
err := anActor.AttemptsTo(
    watch.TheKey("app/config/database"),
    wait.UntilThe(events.OnTheKey("app/config/database")).HappensToHave(length.AtLeast[[]*ability.Event](1)),
)
// Alice attempts to wait until the events on the key "app/config/database" happens to have length at least 1
err = anActor.Should(see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()))
// Alice should see that the last of events on the key "app/config/database" should have been put
```
An `*ability.Event` exposes its parts through getters: `Key()`, `IsPut()` and
`IsDelete()`, `Value()` (empty for a delete), `PreviousValue()` (empty when the event
created the key), `Revision()`, `Version()` (zero after a delete), and
`CausedByLeaseExpiry()`, which tells a key removed because its lease expired from a key an
actor removed:
```go
recorded, _ := anActor.AsksFor(events.OnTheKey("locks/batch-job")) // recorded is a []*ability.Event

for _, event := range recorded {
    if event.IsDelete() && event.CausedByLeaseExpiry() {
        fmt.Printf("%s expired at revision %d, it held %q\n", event.Key(), event.Revision(), event.PreviousValue())
    }
}
```

### The leader of an election
`leader.OfTheElection` answers with the value the leader of the election was announced
with, a `string`: the name of the actor leading it, or the value it campaigned with. An
election without a leader answers with an empty value, so the question never fails on
that account:
```go
err := anActor.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("Alice")))
// Alice should see that the leader of the election "primary-worker" should be equal to "Alice"
err := anActor.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.Empty[string]()))
```

### The health of the cluster
Every scenario starts with a running cluster. Rather than letting the first put report
that it is not, ask for `status.OfTheEndpoint`: it answers with the `*ability.EndpointStatus`
of the member listening on the endpoint, and fails when the member cannot be reached. The
status exposes `Version()`, `Leader()` (the ID of the member believed to lead the cluster,
zero when none is known), `DBSize()` in bytes, `IsLearner()`, and `Errors()`, the alarms
and health problems the member reports:
```go
member, err := anActor.AsksFor(status.OfTheEndpoint("localhost:2379")) // member is an *ability.EndpointStatus
// Alice asks for the status of the endpoint "localhost:2379"
err = anActor.Should(see.The(status.OfTheEndpoint("localhost:2379")).ShouldNotBe(be.Zero[*ability.EndpointStatus]()))
```
`alarms.OfTheCluster` answers with the alarms raised by the members of the cluster, a
`[]*ability.Alarm` with `MemberID()` and `Type()` (`NOSPACE` or `CORRUPT`), empty on a
healthy cluster:
```go
err := anActor.Should(see.The(alarms.OfTheCluster()).Should(be.Empty[[]*ability.Alarm]()))
// Alice should see that the alarms of the cluster should be empty
```

## Resolutions

### Existence
`exist.InETCD` tells if the key answered by `key.Named` exists. It is a bare infinitive
and takes the verb `Should` or `ShouldNot`; a nil key does not exist:
```go
err := anActor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
// Alice should see that the key named "app/status" should exist in etcd
```
### Events
The `been` resolutions tell what an event did to its key. They are past participles and
take the verb `ShouldHave` or `ShouldNotHave`. `been.Put()` matches a put, and
`been.Deleted()` a delete, whatever its cause; close the latter with
`ByTheExpiryOfItsLease` to require the key to have expired rather than been removed by an
actor:
```go
err := anActor.Should(see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()))
// Alice should see that the last of events on the key "app/config/database" should have been put
err := anActor.Should(see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted()))
// Alice should see that the last of events on the key "locks/batch-job" should have been deleted
err := anActor.Should(see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted().ByTheExpiryOfItsLease()))
// Alice should see that the last of events on the key "locks/batch-job" should have been deleted by the expiry of its lease
// fails: Alice failed to see that the last of events on the key "locks/batch-job" should have been deleted by the expiry of its lease: it has been deleted by an actor, not by the expiry of its lease
```

### Fakes
To test a resolution or an action of your own without a store, the `question/fixture`
package creates fake keys and fake events:
```go
existing := fixture.CreateFakeExistingKey("app/status", "ready")
missing := fixture.CreateFakeMissingKey("does/not/exist")
put := fixture.CreateFakePutEvent("app/status", "ready")
deleted := fixture.CreateFakeDeleteEvent("app/status", "ready")
expired := fixture.CreateFakeExpiryEvent("locks/batch-job", "held")
alarm := fixture.CreateFakeAlarm(memberID, "NOSPACE")
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

`RunSecureETCDServer` starts a server that secures its connections with TLS and requires
its clients to present a certificate. It generates a certificate authority, a server
certificate and a client certificate for the test, so a scenario exercises the secure
settings of the ability without a real cluster:
```go
server := fixture.RunSecureETCDServer(t)
anActor := screenplay.ActorNamed("Alice").WhoCan(
    UseETCD().
        WithTheEndpoints(server.Endpoint()).
        WithTheCertificateAuthority(server.CertificateAuthority()).
        WithTheClientCertificate(server.ClientCertificate()).AndTheKey(server.ClientKey()),
)
```

## Putting it all together
A complete program, from giving an actor the ability to releasing the connection. It
stores a configuration in one transaction, checks what the store holds, removes one key
then everything left under the prefix, and exits:
```go
package main

import (
	"log"

	"github.com/grandper/go-screenplay/v2/action/see"
	"github.com/grandper/go-screenplay/v2/question/number"
	"github.com/grandper/go-screenplay/v2/resolution/be"
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
	"github.com/grandper/go-screenplay-etcd/action/put"
	"github.com/grandper/go-screenplay-etcd/action/remove"
	"github.com/grandper/go-screenplay-etcd/question/key"
	"github.com/grandper/go-screenplay-etcd/question/keys"
	"github.com/grandper/go-screenplay-etcd/question/value"
	"github.com/grandper/go-screenplay-etcd/resolution/exist"
)

func main() {
	alice := screenplay.ActorNamed("Alice").WhoCan(
		ability.UseETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(5).Seconds(),
	)

	// Store the configuration in one transaction.
	err := alice.AttemptsTo(put.TheValues(map[string]string{
		"app/config/database": "postgresql://localhost:5432",
		"app/config/cache":    "redis://localhost:6379",
	}))
	if err != nil {
		log.Fatalf("failed to store the configuration: %v", err)
	}

	// Check what the store holds.
	err = alice.Should(
		see.The(key.Named("app/config/database")).Should(exist.InETCD()),
		see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")),
		see.The(number.Of(keys.StartingWith("app/config/"))).Should(be.EqualTo(2)),
	)
	if err != nil {
		log.Fatalf("the configuration is not as expected: %v", err)
	}

	// Clean up: one key, then everything left under the prefix.
	err = alice.AttemptsTo(
		remove.TheKey("app/config/cache"),
		remove.TheKeysStartingWith("app/config/"),
		see.The(key.Named("app/config/database")).ShouldNot(exist.InETCD()),
	)
	if err != nil {
		log.Fatalf("failed to clean up the configuration: %v", err)
	}

	// Release the connection.
	if err = alice.Exit(); err != nil {
		log.Fatalf("Alice failed to exit: %v", err)
	}
}
```
Every import above is used as shown; the program builds as it is against an etcd server
listening on `localhost:2379`. In a test, replace the endpoint with the one of an embedded
server (see below).

## Limitations
- There is no retry, and no backoff, around the operations of the ability: a failure is
  reported by the action that met it. Wrap an action with the core `keep.TryingTo` to
  retry it.
- A lease lives for whole seconds, which is how etcd counts, and the cluster raises a short
  lease to the minimum it allows (two seconds with the defaults of the embedded server),
  so a key expires a little later than written.
- The ability authenticates with certificates only. For a user name and a password, or for
  any other option of the etcd client, create the client yourself and give it with
  `WithTheClient`.
- Removing only the keys under a prefix whose value satisfies a condition, and putting or
  removing several keys conditionally in one transaction, are not expressed yet: ask for
  `keys.StartingWith` and write one conditional action per key.

## Development
The tests run against the embedded server of the `fixture` package, with the race detector:
```bash
go test -race ./...
golangci-lint run ./...
```
Until go-screenplay v2.0.0 is tagged, `go.mod` replaces `github.com/grandper/go-screenplay/v2`
with the `release/2.0.0` branch checked out next to this repository, in `../go-screenplay`.

## License
This extension is released under the [MIT License](LICENSE).
