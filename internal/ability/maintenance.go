package ability

import (
	"context"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"

	"github.com/grandper/go-screenplay-etcd/internal/answer"
)

// ToRetrieveTheStatusOf retrieves the status of the member of the cluster
// listening on the endpoint.
func (o *Operations) ToRetrieveTheStatusOf(ctx context.Context, endpoint string) (*answer.EndpointStatus, error) {
	client, err := o.client()
	if err != nil {
		return nil, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.Status(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	return answer.NewEndpointStatus(response), nil
}

// ToRetrieveTheAlarms retrieves the alarms raised by the members of the
// cluster, none on a healthy cluster.
func (o *Operations) ToRetrieveTheAlarms(ctx context.Context) ([]*answer.Alarm, error) {
	client, err := o.client()
	if err != nil {
		return nil, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	list, err := client.AlarmList(ctx)
	if err != nil {
		return nil, err
	}

	response := (*pb.AlarmResponse)(list)

	alarms := make([]*answer.Alarm, 0, len(response.GetAlarms()))
	for _, alarm := range response.GetAlarms() {
		alarms = append(alarms, answer.NewAlarm(alarm.GetMemberID(), alarm.GetAlarm().String()))
	}

	return alarms, nil
}
