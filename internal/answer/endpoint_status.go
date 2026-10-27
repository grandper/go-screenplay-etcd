package answer

import (
	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EndpointStatus is the status of a member of the cluster, as the actor saw
// it. Its parts are read through its accessors.
type EndpointStatus struct {
	version   string
	leader    uint64
	dbSize    int64
	isLearner bool
	errors    []string
}

// NewEndpointStatus creates the status a member answered with.
func NewEndpointStatus(status *clientv3.StatusResponse) *EndpointStatus {
	response := (*pb.StatusResponse)(status)

	return &EndpointStatus{
		version:   response.GetVersion(),
		leader:    response.GetLeader(),
		dbSize:    response.GetDbSize(),
		isLearner: response.GetIsLearner(),
		errors:    append([]string{}, response.GetErrors()...),
	}
}

// Version returns the version of etcd the member runs.
func (s *EndpointStatus) Version() string {
	return s.version
}

// Leader returns the ID of the member the member believes is the leader of
// the cluster, zero when it knows of none.
func (s *EndpointStatus) Leader() uint64 {
	return s.leader
}

// DBSize returns the size of the database of the member, in bytes.
func (s *EndpointStatus) DBSize() int64 {
	return s.dbSize
}

// IsLearner returns true when the member is a learner, which does not vote.
func (s *EndpointStatus) IsLearner() bool {
	return s.isLearner
}

// Errors returns the alarms and the health problems the member reports, none
// when it is healthy.
func (s *EndpointStatus) Errors() []string {
	return append([]string{}, s.errors...)
}

// Alarm is an alarm raised by a member of the cluster, such as NOSPACE when it
// runs out of space.
type Alarm struct {
	memberID  uint64
	alarmType string
}

// NewAlarm creates an alarm raised by the member.
func NewAlarm(memberID uint64, alarmType string) *Alarm {
	return &Alarm{memberID: memberID, alarmType: alarmType}
}

// MemberID returns the ID of the member that raised the alarm.
func (a *Alarm) MemberID() uint64 {
	return a.memberID
}

// Type returns the type of the alarm: NOSPACE or CORRUPT.
func (a *Alarm) Type() string {
	return a.alarmType
}
