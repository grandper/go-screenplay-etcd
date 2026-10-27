// Package alarms holds the question that answers with the alarms raised by
// the members of the cluster:
//
//	err := anActor.Should(see.The(alarms.OfTheCluster()).Should(be.Empty[[]*etcd.Alarm]()))
//
// reads "see the alarms of the cluster should be empty".
package alarms
