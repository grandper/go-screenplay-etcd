// Package ability holds the implementation of the ability to use etcd. Only
// the actions and the questions of the extension use its operations, through
// TheAbilityOf; a test gives the ability to an actor through the public
// package:
//
//	anActor := screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints("localhost:2379"))
package ability
