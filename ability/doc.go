// Package ability holds the ability to use etcd:
//
//	anActor := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints("localhost:2379"))
//
// The ability is configured fluently, reading "use etcd with the endpoints
// localhost:2379 with the namespace test-run-42/ with timing out after 5
// seconds":
//
//	anActor := screenplay.ActorNamed("Alice").WhoCan(
//		ability.UseETCD().
//			WithTheEndpoints("localhost:2379").
//			WithTheNamespace("test-run-42/").
//			WithTimingOutAfter(5).Seconds(),
//	)
package ability
