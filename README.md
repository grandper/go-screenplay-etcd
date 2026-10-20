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
| `fixture` | `github.com/grandper/go-screenplay-etcd/fixture` | `RunETCDServer`, `ETCDServer` |

## Testing with an embedded etcd server
The `fixture` package spins up an in-process etcd server, so scenarios run without an
external cluster and without Docker. `RunETCDServer` takes a `*testing.T` and registers
its own cleanup, so the server is stopped when the test ends. Every call starts its own
server on a free port of the loopback interface, so tests never share keys:
```go
server := fixture.RunETCDServer(t)

endpoint := server.Endpoint() // "127.0.0.1:52817"
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
