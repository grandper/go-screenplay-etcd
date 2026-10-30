// Package use holds the ability to use etcd, given to an actor with WhoCan:
//
//	anActor := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints("localhost:2379"))
//
// The ability is configured fluently, reading "use etcd with the endpoints
// localhost:2379 with the namespace test-run-42/ with timing out after 5
// seconds":
//
//	anActor := screenplay.ActorNamed("Alice").WhoCan(
//		use.ETCD().
//			WithTheEndpoints("localhost:2379").
//			WithTheNamespace("test-run-42/").
//			WithTimingOutAfter(5).Seconds(),
//	)
//
// A client the test created is given with WithTheClient instead of the
// endpoints. A cluster secured with TLS is given the authority that signed its
// certificates with WithTheCertificateAuthority, and the certificate the actor
// presents with WithTheClientCertificate closed by AndTheKey:
//
//	anActor := screenplay.ActorNamed("Alice").WhoCan(
//		use.ETCD().
//			WithTheEndpoints("etcd.example.com:2379").
//			WithTheCertificateAuthority("ca.crt").
//			WithTheClientCertificate("client.crt").AndTheKey("client.key"),
//	)
//
// What the ability answers with, such as a key or an event, and the errors it
// reports live in the etcd package.
package use
