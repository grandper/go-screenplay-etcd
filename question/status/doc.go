// Package status holds the question that answers with the status of a member
// of the cluster:
//
//	status, err := anActor.AsksFor(status.OfTheEndpoint("localhost:2379")) // status is an *ability.EndpointStatus
//
// reads "asks for the status of the endpoint localhost:2379".
package status
