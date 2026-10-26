package etcd

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

// EndpointStatus is the status of a member of the cluster. Its parts are read
// through Version(), Leader(), DBSize(), IsLearner() and Errors().
type EndpointStatus = impl.EndpointStatus

// Alarm is an alarm raised by a member of the cluster. Its parts are read
// through MemberID() and Type().
type Alarm = impl.Alarm
