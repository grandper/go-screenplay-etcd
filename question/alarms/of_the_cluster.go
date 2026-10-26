package alarms

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/etcd"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
)

// OfTheCluster asks about the alarms raised by the members of the cluster,
// such as NOSPACE when a member runs out of space. A healthy cluster answers
// with no alarm:
//
//	err := anActor.Should(see.The(alarms.OfTheCluster()).Should(be.Empty[[]*etcd.Alarm]()))
func OfTheCluster() screenplay.Question[[]*etcd.Alarm] {
	return &ofTheClusterQuestion{}
}

// ofTheClusterQuestion asks about the alarms raised by the cluster.
type ofTheClusterQuestion struct{}

// AnsweredBy returns the alarms raised by the members of the cluster.
func (q *ofTheClusterQuestion) AnsweredBy(theActor *screenplay.Actor) ([]*etcd.Alarm, error) {
	return impl.TheAbilityOf(theActor).ToRetrieveTheAlarms(theActor.Context())
}

// String describes the question.
func (q *ofTheClusterQuestion) String() string {
	return "alarms of the cluster"
}

// ofTheClusterQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]*etcd.Alarm] = (*ofTheClusterQuestion)(nil)
