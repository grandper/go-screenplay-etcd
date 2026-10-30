# go-screenplay-etcd

[![Test](https://github.com/grandper/go-screenplay-etcd/actions/workflows/go-test.yml/badge.svg)](https://github.com/grandper/go-screenplay-etcd/actions/workflows/go-test.yml)
[![Lint](https://github.com/grandper/go-screenplay-etcd/actions/workflows/go-lint.yml/badge.svg)](https://github.com/grandper/go-screenplay-etcd/actions/workflows/go-lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/grandper/go-screenplay-etcd.svg)](https://pkg.go.dev/github.com/grandper/go-screenplay-etcd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

This project is an extension of [go-screenplay](https://github.com/grandper/go-screenplay) to drive an [etcd](https://etcd.io) cluster. It bundles the ability to use etcd together with the actions, questions and resolutions that go with it, so you test a system built on etcd with the same screenplay vocabulary you use everywhere else. A test reads as English, and everything the actor does is narrated:
```go
actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints("localhost:2379"))

err := actor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// Alice attempts to put the value "ready" for the key "app/status"
err = actor.Should(see.The(value.OfTheKey("app/status")).Should(be.EqualTo("ready")))
// Alice should see that the value of the key "app/status" should be equal to "ready"
```

## Installation
```sh
go get github.com/grandper/go-screenplay-etcd
```
It requires Go 1.27 or later and go-screenplay v2, and drives etcd v3 through its official client, `go.etcd.io/etcd/client/v3`.

## Ability
This extension provides one ability, `use.ETCD()`, which allows an actor to interact with an etcd cluster; given to an actor with `WhoCan`, it reads "who can use etcd". The ability is given the endpoints of the cluster with `WithTheEndpoints("localhost:2379")`, reading "use etcd with the endpoints localhost:2379", or a client of your own with `WithTheClient(client)`, reading "use etcd with the client".

First, import the package:
```go
import (
    clientv3 "go.etcd.io/etcd/client/v3"
    "github.com/grandper/go-screenplay/v2/screenplay"
    "github.com/grandper/go-screenplay/v2/action/keep"
    "github.com/grandper/go-screenplay/v2/action/see"
    "github.com/grandper/go-screenplay/v2/action/wait"
    "github.com/grandper/go-screenplay/v2/question/last"
    "github.com/grandper/go-screenplay/v2/question/number"
    "github.com/grandper/go-screenplay/v2/resolution/be"
    "github.com/grandper/go-screenplay/v2/resolution/contain"
    "github.com/grandper/go-screenplay/v2/resolution/length"
    "github.com/grandper/go-screenplay-etcd/ability/use"
    "github.com/grandper/go-screenplay-etcd/action/acquire"
    "github.com/grandper/go-screenplay-etcd/action/campaign"
    "github.com/grandper/go-screenplay-etcd/action/put"
    "github.com/grandper/go-screenplay-etcd/action/release"
    "github.com/grandper/go-screenplay-etcd/action/remove"
    "github.com/grandper/go-screenplay-etcd/action/resign"
    "github.com/grandper/go-screenplay-etcd/action/watch"
    "github.com/grandper/go-screenplay-etcd/etcd"
    "github.com/grandper/go-screenplay-etcd/question/alarms"
    "github.com/grandper/go-screenplay-etcd/question/events"
    "github.com/grandper/go-screenplay-etcd/question/key"
    "github.com/grandper/go-screenplay-etcd/question/keys"
    "github.com/grandper/go-screenplay-etcd/question/leader"
    "github.com/grandper/go-screenplay-etcd/question/revision"
    "github.com/grandper/go-screenplay-etcd/question/status"
    "github.com/grandper/go-screenplay-etcd/question/ttl"
    "github.com/grandper/go-screenplay-etcd/question/value"
    "github.com/grandper/go-screenplay-etcd/question/values"
    "github.com/grandper/go-screenplay-etcd/resolution/been"
    "github.com/grandper/go-screenplay-etcd/resolution/exist"
)
```
Then create an actor who can use etcd, with the endpoints of the cluster in the form `host:port`:
```go
actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints("localhost:2379"))
actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints("etcd-1:2379", "etcd-2:2379", "etcd-3:2379"))
```
The connection is opened the first time the actor uses the ability. An unreachable cluster makes the first action wait for the timeout of the ability (see `WithTimingOutAfter` below) or for the deadline of the actor's context, and fail with `context.DeadlineExceeded`; give the ability one of the two, otherwise the action waits for the cluster indefinitely. An actor given no endpoint fails every action with `etcd.ErrNoEndpointProvided`:
```go
actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD())

err := actor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// no etcd endpoint was provided to the ability
```
Alternatively, give the ability a client you created yourself, for example to authenticate with a user name and a password, or to use an option of the etcd client the ability does not expose. The ability owns the client from then on:
```go
client, _ := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}, Username: "alice", Password: "s3cret"})

actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheClient(client))
```
The ability is configured fluently, and the chain reads as English: "use etcd with the endpoints localhost:2379 with the namespace test-run-42/ with timing out after 5 seconds":
```go
actor := screenplay.ActorNamed("Alice").WhoCan(
    use.ETCD().
        WithTheEndpoints("localhost:2379").
        WithTheNamespace("test-run-42/").
        WithTimingOutAfter(5).Seconds(),
)
```
`WithTheNamespace` prefixes every key the actor uses. The actions and the questions keep using plain keys: the prefix is added on the wire and removed from what the cluster answers, so concurrent scenarios run against one cluster without sharing keys:
```go
err := actor.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"))
// the cluster holds the key "test-run-42/app/config/database"
```
`WithTimingOutAfter` bounds every operation of the actor; without it, an operation is bounded only by the deadline of the actor's context (see `screenplay.WithContext`). The timeout is closed by its unit, `Milliseconds`, `Seconds` or `Minutes`, each with its singular form so that one unit reads as English. An operation that outlasts the timeout fails with `context.DeadlineExceeded`:
```go
actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(500).Milliseconds())
actor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(1).Minute())
```
A cluster that secures its connections with TLS is given the authority that signed its certificates with `WithTheCertificateAuthority`; without it, the authorities of the system are trusted. A cluster that authenticates its clients with mutual TLS is also given the certificate the actor presents, with `WithTheClientCertificate` closed by `AndTheKey`. All three are PEM files:
```go
actor := screenplay.ActorNamed("Alice").WhoCan(
    use.ETCD().
        WithTheEndpoints("etcd.example.com:2379").
        WithTheCertificateAuthority("ca.crt").
        WithTheClientCertificate("client.crt").AndTheKey("client.key"),
)
```
A file that cannot be read, or that holds no certificate, is reported by the first action, with an error wrapping both `etcd.ErrFailedToConnect` and `etcd.ErrFailedToLoadTheCertificates`:
```go
err := actor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// failed to connect to the etcd server: failed to load the certificates: open ca.crt: no such file or directory
```

> **_NOTE:_** When the actor exits (`actor.Exit()`), the ability forgets what it holds: the watches it started and the renewals of the leases it keeps alive stop, the locks it holds are released, the elections it campaigns in are resigned from, and the connection is closed, whether it was opened by the ability or given with `WithTheClient`.

The settings are the only methods a test can call on the ability: everything the actor does with etcd goes through the actions and the questions of this package. An actor that was not given the ability fails every action and every question with `screenplay.ErrActorMissingAbility`:
```go
actor := screenplay.ActorNamed("Bob")

err := actor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// Bob attempts to put the value "ready" for the key "app/status"
// fails: actor does not have the required ability
```

You can of course use all the available options when creating the etcd client yourself, see the [etcd Go client](https://github.com/etcd-io/etcd/tree/main/client/v3) for more details.

## Actions
The actions live in packages named after their verb, so that a statement reads as English: `put.TheValue("ready").ForTheKey("app/status")` reads "put the value ready for the key app/status". Every action is narrated word for word as it was written, after the name of the actor:
```go
err := actor.AttemptsTo(put.TheValue("ready").ForTheKey("app/status"))
// Alice attempts to put the value "ready" for the key "app/status"
```
Nothing panics when an action is built. A mistake, such as an empty key, is reported when the actor performs the action, with an error of the `etcd` package: `etcd.ErrKeyNotProvided`, `etcd.ErrPrefixNotProvided` or `etcd.ErrNameNotProvided`.

The core actions compose with the ones of this extension as with any other: `keep.TryingTo` retries a put against a cluster that is still starting, `proceed.ConcurrentlyTo` runs several actions at once, and `wait.UntilThe` waits for a question of this extension to answer as expected:
```go
err := actor.AttemptsTo(keep.TryingTo(put.TheValue("ready").ForTheKey("app/status")).For(30).Seconds())
// Alice attempts to keep trying to put the value "ready" for the key "app/status" for 30 seconds
```

### put.TheValue
Puts a value for the key given with `ForTheKey`. The key is created when it does not exist, and its value is overwritten when it does:
```go
err := actor.AttemptsTo(put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"))
// Alice attempts to put the value "postgresql://localhost:5432" for the key "app/config/database"
```
It fails with `etcd.ErrKeyNotProvided` when the key is missing or empty:
```go
err := actor.AttemptsTo(put.TheValue("postgresql://localhost:5432"))
// the key must be provided
```
A value you do not want in the narration, such as a password, is kept out of it with `Secretly` (or its longer form `WhichShouldBeKeptSecret`):
```go
err := actor.AttemptsTo(put.TheValue(password).ForTheKey("app/config/password").Secretly())
err := actor.AttemptsTo(put.TheValue(password).ForTheKey("app/config/password").WhichShouldBeKeptSecret())
// Alice attempts to put the value <secret> for the key "app/config/password"
```
A key can be given a time to live with `ExpiringAfter` and a unit (`Seconds`, `Minutes` or `Milliseconds`, with their singular forms): the value is put with a lease, and the cluster removes the key when the lease expires. etcd counts in whole seconds, so the time is rounded up, and a cluster raises a short lease to the minimum it allows (two seconds with the default settings of etcd):
```go
err := actor.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(10).Seconds())
// Alice attempts to put the value "held" for the key "locks/batch-job" expiring after 10 seconds
```
A long-running actor, such as a worker holding a soft lock, renews the lease instead of letting it expire. `AndKeepItAlive` renews it in the background for as long as the actor keeps the ability, so the key disappears only once the actor exits:
```go
err := actor.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(10).Seconds().AndKeepItAlive())
// Alice attempts to put the value "held" for the key "locks/batch-job" expiring after 10 seconds and keep it alive

err = actor.Exit() // the renewals stop, and the key expires ten seconds later
```
To initialize a key only when no one else did, close the statement with `IfTheKeyIsAbsent`: the check and the put happen in one transaction, and a key that exists makes the action fail with an error wrapping `etcd.ErrKeyAlreadyExists`, leaving the cluster untouched:
```go
err := actor.AttemptsTo(put.TheValue("initial").ForTheKey("app/config/flag").IfTheKeyIsAbsent())
// Alice attempts to put the value "initial" for the key "app/config/flag" if the key is absent
// fails: the key "app/config/flag" already exists
```
To update a key only when nobody changed it since you read it, ask for its revision with `revision.OfTheKey`, then close the statement with `IfTheKeyIsUnchangedSinceTheRevision`. A key modified in between makes the action fail with an error wrapping `etcd.ErrKeyModified`, so a read-modify-write loop knows to read again. The revision zero means that the key must not exist, which is another way of saying `IfTheKeyIsAbsent`:
```go
lastModification, _ := actor.AsksFor(revision.OfTheKey("app/config/counter"))

err := actor.AttemptsTo(put.TheValue("2").ForTheKey("app/config/counter").IfTheKeyIsUnchangedSinceTheRevision(lastModification))
// Alice attempts to put the value "2" for the key "app/config/counter" if the key is unchanged since the revision 42
// fails: the key "app/config/counter" was modified since the revision 42
```
The expiry and the conditions combine, and are narrated in the order they are written: a soft lock is acquired by putting a value expiring after a while, if the key is absent:
```go
err := actor.AttemptsTo(put.TheValue("held").ForTheKey("locks/batch-job").ExpiringAfter(10).Seconds().IfTheKeyIsAbsent())
// Alice attempts to put the value "held" for the key "locks/batch-job" expiring after 10 seconds if the key is absent
```

### put.TheValues
Puts several values, each for its key, in one transaction, so the cluster never holds some of them without the others. The narration lists them sorted by key:
```go
err := actor.AttemptsTo(put.TheValues(map[string]string{
    "app/config/database": "postgresql://localhost:5432",
    "app/config/cache":    "redis://localhost:6379",
}))
// Alice attempts to put the values "app/config/cache" = "redis://localhost:6379", "app/config/database" = "postgresql://localhost:5432"
```
It fails with `etcd.ErrKeyNotProvided` when the map is empty or holds an empty key. The values are kept secret the same way as a single one; the keys are still narrated:
```go
err := actor.AttemptsTo(put.TheValues(credentials).Secretly())
// Alice attempts to put the values <secret> for the keys "app/config/password", "app/config/username"
```

### remove.TheKey
Removes a key. Removing a key that does not exist is not an error, so a cleanup task never fails for a key a scenario did not create:
```go
err := actor.AttemptsTo(remove.TheKey("app/config/cache"))
// Alice attempts to remove the key "app/config/cache"
```
The key is removed only if nobody changed it since a revision with `IfUnchangedSinceTheRevision`, which fails with an error wrapping `etcd.ErrKeyModified` as the conditional put does:
```go
err := actor.AttemptsTo(remove.TheKey("app/config/counter").IfUnchangedSinceTheRevision(lastModification))
// Alice attempts to remove the key "app/config/counter" if unchanged since the revision 42
```
`remove.TheKeys` removes several keys in one transaction, so the cluster never holds some of them without the others:
```go
err := actor.AttemptsTo(remove.TheKeys("app/config/cache", "app/config/database"))
// Alice attempts to remove the keys "app/config/cache", "app/config/database"
```
Both fail with `etcd.ErrKeyNotProvided` when a key is empty, or when no key is given at all:
```go
err := actor.AttemptsTo(remove.TheKeys())
// the key must be provided
```

### remove.TheKeysStartingWith
Removes every key starting with a prefix:
```go
err := actor.AttemptsTo(remove.TheKeysStartingWith("app/config/"))
// Alice attempts to remove the keys starting with "app/config/"
```
It fails with `etcd.ErrPrefixNotProvided` when the prefix is empty, which would reach every key of the cluster:
```go
err := actor.AttemptsTo(remove.TheKeysStartingWith(""))
// the prefix must be provided
```

### watch.TheKey
Watches a key: every change made to it from then on is recorded, until the actor exits, where `events.OnTheKey(key)` finds it. The action returns once the cluster confirmed the watch, so a change made right after it is never missed:
```go
err := actor.AttemptsTo(watch.TheKey("app/config/database"))
// Alice attempts to watch the key "app/config/database"
```
`watch.TheKeysStartingWith` watches every key starting with a prefix, where `events.OnTheKeysStartingWith(prefix)` finds the changes:
```go
err := actor.AttemptsTo(watch.TheKeysStartingWith("app/config/"))
// Alice attempts to watch the keys starting with "app/config/"
```
Watching twice the same key, or the same prefix, fails with an error wrapping `etcd.ErrAlreadyWatching`; a watch of a key and a watch of the keys starting with it are two different watches:
```go
err := actor.AttemptsTo(watch.TheKey("app/config/database"), watch.TheKey("app/config/database"))
// the key "app/config/database" is already watched
```

### acquire.TheLock
Acquires the distributed lock of the given name, which only one actor at a time may hold, waiting for the actor holding it to release it; the wait ends with the context of the actor (see `screenplay.WithContext`):
```go
err := actor.AttemptsTo(acquire.TheLock("deployment"))
// Alice attempts to acquire the lock "deployment"
```
To fail instead of waiting when another actor holds the lock, close the statement with `WithoutWaiting`; the action then fails with `etcd.ErrLocked`, the error of the etcd client:
```go
err := anotherActor.AttemptsTo(acquire.TheLock("deployment").WithoutWaiting())
// Bob attempts to acquire the lock "deployment" without waiting
// fails: mutex: Locked by another session
```
Acquiring a lock the actor already holds fails with an error wrapping `etcd.ErrLockAlreadyHeld`, and an empty name with `etcd.ErrNameNotProvided`. The locks of an actor share one session, a lease the client keeps alive: should the actor lose its connection to the cluster for longer than the session allows, the cluster releases its locks.

### release.TheLock
Releases a lock the actor holds; a lock is also released when the actor exits:
```go
err := actor.AttemptsTo(release.TheLock("deployment"))
// Alice attempts to release the lock "deployment"
```
It fails with an error wrapping `etcd.ErrLockNotHeld` when the actor does not hold the lock:
```go
err := actor.AttemptsTo(release.TheLock("deployment"))
// the lock "deployment" is not held by the actor
```

### campaign.InTheElection
Campaigns in the election of the given name, which names a leader among the actors campaigning in it, and returns once the actor leads it, waiting for the leader to resign if there is one. The actor is announced under its name, unless it campaigns `WithTheValue` of its own:
```go
err := actor.AttemptsTo(campaign.InTheElection("primary-worker"))
// Alice attempts to campaign in the election "primary-worker"
err := actor.AttemptsTo(campaign.InTheElection("primary-worker").WithTheValue("node-1"))
// Alice attempts to campaign in the election "primary-worker" with the value "node-1"
```
Campaigning in an election the actor already campaigns in fails with an error wrapping `etcd.ErrAlreadyCampaigning`, and an empty name with `etcd.ErrNameNotProvided`.

### resign.FromTheElection
Gives up the leadership of an election the actor campaigns in; the actor also resigns when it exits:
```go
err := actor.AttemptsTo(resign.FromTheElection("primary-worker"))
// Alice attempts to resign from the election "primary-worker"
```
It fails with an error wrapping `etcd.ErrNotCampaigning` when the actor does not campaign in the election:
```go
err := actor.AttemptsTo(resign.FromTheElection("primary-worker"))
// the actor does not campaign in the election "primary-worker"
```

## Questions
A question is a noun phrase that the actor asks for, and its package is its first word: `value.OfTheKey("app/status")` reads "value of the key app/status". Its answer is typed: `see.The`, `AsksFor` and the core questions such as `last.Of` and `number.Of` work on it without any type assertion.

### value.OfTheKey
`value.OfTheKey(key)` answers with the value the key holds, a `string`, so you assert on it like any text:
```go
err := actor.Should(see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")))
// Alice should see that the value of the key "app/config/database" should be equal to "postgresql://localhost:5432"
databaseURL, err := actor.AsksFor(value.OfTheKey("app/config/database")) // databaseURL is a string
```
It fails with an error wrapping `etcd.ErrKeyNotFound` when the key does not exist, and with `etcd.ErrKeyNotProvided` when the key is empty:
```go
missing, err := actor.AsksFor(value.OfTheKey("does/not/exist"))
// the key "does/not/exist" was not found
```

### key.Named
`key.Named(key)` answers with an `*etcd.Key`, whether the key exists or not, so you assert on its existence with the `exist.InETCD` resolution:
```go
err := actor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
// Alice should see that the key named "app/status" should exist in etcd
err := actor.Should(see.The(key.Named("app/status")).ShouldNot(exist.InETCD()))
// Alice should see that the key named "app/status" should not exist in etcd
```
An `*etcd.Key` is read through `Name()` and `Exists()`, then `Value()`, `Version()` (the number of times the key was put since it was created), `CreateRevision()` and `ModRevision()` (the revisions of the cluster at which it was created and last modified), and `Lease()` (the ID of the lease attached to it, zero without a lease). They are all zero or empty when the key does not exist:
```go
database, _ := actor.AsksFor(key.Named("app/config/database")) // database is an *etcd.Key

if database.Exists() && database.Version() > 1 {
    fmt.Printf("%s was updated %d times, last at revision %d\n", database.Name(), database.Version()-1, database.ModRevision())
}
```

### revision.OfTheKey
`revision.OfTheKey(key)` answers with the revision at which the key was last modified, an `int64`, which is what a conditional write needs. It fails like `value.OfTheKey` when the key does not exist:
```go
lastModification, err := actor.AsksFor(revision.OfTheKey("app/config/database")) // lastModification is an int64
err = actor.Should(see.The(revision.OfTheKey("app/config/database")).Should(be.GreaterThan(lastModification)))
// Alice should see that the revision of the key "app/config/database" should be greater than 42
```

### ttl.OfTheKey
`ttl.OfTheKey(key)` answers with the time to live of the key, the time it has left before its lease expires and removes it, a `time.Duration` in whole seconds. A key without a lease never expires, and has zero time to live. It fails like `value.OfTheKey` when the key does not exist:
```go
err := actor.Should(see.The(ttl.OfTheKey("locks/batch-job")).Should(be.AtMost(10*time.Second)))
// Alice should see that the TTL of the key "locks/batch-job" should be at most 10s
err := actor.Should(see.The(ttl.OfTheKey("app/config/database")).Should(be.Zero[time.Duration]()))
```

### keys.StartingWith
`keys.StartingWith(prefix)` answers with the keys starting with the prefix, a `[]string` in the order of the keys, and `values.OfTheKeysStartingWith(prefix)` with a `map[string]string` from each key to its value. The keys are counted with the core `number.Of`:
```go
err := actor.Should(see.The(keys.StartingWith("app/config/")).Should(contain.Element("app/config/database")))
// Alice should see that the keys starting with "app/config/" should contain element "app/config/database"
err := actor.Should(see.The(values.OfTheKeysStartingWith("app/config/")).Should(contain.Key[string, string]("app/config/database")))
// Alice should see that the values of the keys starting with "app/config/" should contain key "app/config/database"
err := actor.Should(see.The(number.Of(keys.StartingWith("app/config/"))).Should(be.EqualTo(2)))
// Alice should see that the number of keys starting with "app/config/" should be equal to 2
```
When no key starts with the prefix, the slice and the map are empty: neither question fails. Both fail with `etcd.ErrPrefixNotProvided` when the prefix is empty:
```go
err := actor.Should(see.The(values.OfTheKeysStartingWith("app/config/")).Should(be.Empty[map[string]string]()))
```

### events.OnTheKey
`events.OnTheKey(key)` answers with the changes recorded on a key the actor watches, a `[]*etcd.Event` in the order they happened, and `events.OnTheKeysStartingWith(prefix)` with the changes recorded on a watched prefix. Both answer with an empty slice when nothing happened yet, and compose with the core `wait.UntilThe`, `length`, `first` and `last` packages, and with the `been` resolutions of this extension:
```go
err := actor.AttemptsTo(
    watch.TheKey("app/config/database"),
    wait.UntilThe(events.OnTheKey("app/config/database")).HappensToHave(length.AtLeast[[]*etcd.Event](1)),
)
// Alice attempts to wait until the events on the key "app/config/database" happens to have length at least 1
err = actor.Should(see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()))
// Alice should see that the last of events on the key "app/config/database" should have been put
```
They fail with an error wrapping `etcd.ErrNotWatching` when the actor does not watch the key or the prefix:
```go
recorded, err := actor.AsksFor(events.OnTheKey("app/config/cache"))
// the key "app/config/cache" is not watched
```
An `*etcd.Event` is read through `Key()`, `IsPut()` and `IsDelete()`, `Value()` (empty for a delete), `PreviousValue()` (empty when the event created the key), `Revision()`, `Version()` (zero after a delete), and `CausedByLeaseExpiry()`, which tells a key removed because its lease expired from a key an actor removed:
```go
recorded, _ := actor.AsksFor(events.OnTheKey("locks/batch-job")) // recorded is a []*etcd.Event

for _, event := range recorded {
    if event.IsDelete() && event.CausedByLeaseExpiry() {
        fmt.Printf("%s expired at revision %d, it held %q\n", event.Key(), event.Revision(), event.PreviousValue())
    }
}
```

### leader.OfTheElection
`leader.OfTheElection(name)` answers with the value the leader of the election was announced with, a `string`: the name of the actor leading it, or the value it campaigned with. An election without a leader answers with an empty value, so the question never fails on that account:
```go
err := actor.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.EqualTo("Alice")))
// Alice should see that the leader of the election "primary-worker" should be equal to "Alice"
err := actor.Should(see.The(leader.OfTheElection("primary-worker")).Should(be.Empty[string]()))
```

### status.OfTheEndpoint
`status.OfTheEndpoint(endpoint)` answers with the `*etcd.EndpointStatus` of the member of the cluster listening on the endpoint, the way to check that the cluster is up before a scenario starts, and fails when the member cannot be reached. The status is read through `Version()`, `Leader()` (the ID of the member believed to lead the cluster, zero when none is known), `DBSize()` in bytes, `IsLearner()`, and `Errors()`, the alarms and health problems the member reports:
```go
member, err := actor.AsksFor(status.OfTheEndpoint("localhost:2379")) // member is an *etcd.EndpointStatus
// Alice asks for the status of the endpoint "localhost:2379"
err = actor.Should(see.The(status.OfTheEndpoint("localhost:2379")).ShouldNot(be.Zero[*etcd.EndpointStatus]()))
// Alice should see that the status of the endpoint "localhost:2379" should not be zero
```

### alarms.OfTheCluster
`alarms.OfTheCluster()` answers with the alarms raised by the members of the cluster, a `[]*etcd.Alarm` read through `MemberID()` and `Type()` (`NOSPACE` or `CORRUPT`), empty on a healthy cluster:
```go
err := actor.Should(see.The(alarms.OfTheCluster()).Should(be.Empty[[]*etcd.Alarm]()))
// Alice should see that the alarms of the cluster should be empty
```

## Resolutions
The `exist` and `been` packages hold the resolutions of this extension. Each one reads after its verb, and a negation lives only in the verb: `ShouldNot(exist.InETCD())` reads "should not exist in etcd".

`exist.InETCD()` inspects the answer of `key.Named(key)` after the verb `Should`. A nil key does not exist:
```go
err := actor.Should(see.The(key.Named("app/status")).Should(exist.InETCD()))
// Alice should see that the key named "app/status" should exist in etcd
// fails: Alice failed to see that the key named "app/status" should exist in etcd: it does not exist
err := actor.Should(see.The(key.Named("app/status")).ShouldNot(exist.InETCD()))
// Alice should see that the key named "app/status" should not exist in etcd
```
`been.Put()` and `been.Deleted()` inspect one answer of `events.OnTheKey(key)` or `events.OnTheKeysStartingWith(prefix)`, after the verb `ShouldHave`. `been.Deleted()` matches a delete whatever its cause; close it with `ByTheExpiryOfItsLease` to require the key to have expired rather than been removed by an actor:
```go
err := actor.Should(see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()))
// Alice should see that the last of events on the key "app/config/database" should have been put
err := actor.Should(see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted()))
// Alice should see that the last of events on the key "locks/batch-job" should have been deleted
err := actor.Should(see.The(last.Of(events.OnTheKey("locks/batch-job"))).ShouldHave(been.Deleted().ByTheExpiryOfItsLease()))
// Alice should see that the last of events on the key "locks/batch-job" should have been deleted by the expiry of its lease
// fails: Alice failed to see that the last of events on the key "locks/batch-job" should have been deleted by the expiry of its lease: it has been deleted by an actor, not by the expiry of its lease
```

## Putting it all together
A complete scenario against a cluster listening on `localhost:2379`, from watching a key to asserting on the events the actor recorded. The namespace keeps the keys of the test apart from the ones of the other tests, and the actor exits when the test ends, so that its watch stops and its connection is closed:
```go
func TestNoticesAConfigurationChange(t *testing.T) {
    actor := screenplay.ActorNamed("Alice").WhoCan(
        use.ETCD().WithTheEndpoints("localhost:2379").WithTheNamespace(t.Name() + "/"),
    )
    t.Cleanup(func() { require.NoError(t, actor.Exit()) })

    require.NoError(t, actor.AttemptsTo(
        watch.TheKey("app/config/database"),
        put.TheValue("postgresql://localhost:5432").ForTheKey("app/config/database"),
        wait.UntilThe(events.OnTheKey("app/config/database")).HappensToHave(length.AtLeast[[]*etcd.Event](1)),
    ))

    require.NoError(t, actor.Should(
        see.The(key.Named("app/config/database")).Should(exist.InETCD()),
        see.The(value.OfTheKey("app/config/database")).Should(be.EqualTo("postgresql://localhost:5432")),
        see.The(last.Of(events.OnTheKey("app/config/database"))).ShouldHave(been.Put()),
    ))
}
```

## References
- [etcd documentation](https://etcd.io/docs/)
- [etcd Go client](https://github.com/etcd-io/etcd/tree/main/client/v3)

## License
Licensed under MIT License.
