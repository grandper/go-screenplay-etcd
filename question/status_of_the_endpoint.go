package question

import (
	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/verdict"
)

// StatusOfTheEndpoint asks about the status of the member of the cluster
// listening on an endpoint, the way to check that the cluster is up before a
// scenario starts:
//
//	status, err := anActor.AsksFor(question.StatusOfTheEndpoint("localhost:2379")) // status is an *ability.EndpointStatus
//
// The question fails when the endpoint cannot be reached.
func StatusOfTheEndpoint(endpoint string) screenplay.Question[*ability.EndpointStatus] {
	return &statusOfTheEndpointQuestion{endpoint: endpoint}
}

// statusOfTheEndpointQuestion asks about the status of a member of the
// cluster.
type statusOfTheEndpointQuestion struct {
	endpoint string
}

// AnsweredBy returns the status of the member listening on the endpoint.
func (q *statusOfTheEndpointQuestion) AnsweredBy(theActor *screenplay.Actor) (*ability.EndpointStatus, error) {
	if q.endpoint == "" {
		return nil, ErrEndpointNotProvided
	}

	return impl.TheAbilityOf(theActor).ToRetrieveTheStatusOf(theActor.Context(), q.endpoint)
}

// String describes the question.
func (q *statusOfTheEndpointQuestion) String() string {
	return "status of the endpoint " + verdict.Describe(q.endpoint)
}

// statusOfTheEndpointQuestion implements the screenplay.Question interface.
var _ screenplay.Question[*ability.EndpointStatus] = (*statusOfTheEndpointQuestion)(nil)
