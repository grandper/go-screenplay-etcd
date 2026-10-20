package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
)

// ActiveAlarms asks about the alarms raised by the members of the cluster,
// such as NOSPACE when a member runs out of space. A healthy cluster answers
// with no alarm:
//
//	err := anActor.Should(see.The(question.ActiveAlarms()).Should(be.Empty[[]*ability.Alarm]()))
func ActiveAlarms() screenplay.Question[[]*ability.Alarm] {
	return &activeAlarmsQuestion{}
}

// activeAlarmsQuestion asks about the alarms raised by the cluster.
type activeAlarmsQuestion struct{}

// AnsweredBy returns the alarms raised by the members of the cluster.
func (q *activeAlarmsQuestion) AnsweredBy(theActor *screenplay.Actor) ([]*ability.Alarm, error) {
	return impl.TheAbilityOf(theActor).ToRetrieveTheAlarms(theActor.Context())
}

// String describes the question.
func (q *activeAlarmsQuestion) String() string {
	return "active alarms"
}

// activeAlarmsQuestion implements the screenplay.Question interface.
var _ screenplay.Question[[]*ability.Alarm] = (*activeAlarmsQuestion)(nil)
