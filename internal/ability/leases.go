package ability

import (
	"context"
	"math"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ToGrantALease grants a lease living for the duration, rounded up to whole
// seconds, which is how etcd counts. A key attached to the lease is removed
// when the lease expires.
func (o *Operations) ToGrantALease(ctx context.Context, timeToLive time.Duration) (clientv3.LeaseID, error) {
	client, err := o.client()
	if err != nil {
		return clientv3.NoLease, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.Grant(ctx, int64(math.Ceil(timeToLive.Seconds())))
	if err != nil {
		return clientv3.NoLease, err
	}

	return response.ID, nil
}

// ToRetrieveTheTimeToLive retrieves the time the lease has left to live. A
// lease that expired, or that was revoked, has no time left: zero.
func (o *Operations) ToRetrieveTheTimeToLive(ctx context.Context, lease clientv3.LeaseID) (time.Duration, error) {
	client, err := o.client()
	if err != nil {
		return 0, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.TimeToLive(ctx, lease)
	if err != nil {
		return 0, err
	}

	if response.TTL < 0 {
		return 0, nil
	}

	return time.Duration(response.TTL) * time.Second, nil
}

// ToKeepTheLeaseAlive renews the lease in the background, for as long as the
// ability is not forgotten: when the actor exits, the renewals stop and the
// lease expires.
func (o *Operations) ToKeepTheLeaseAlive(lease clientv3.LeaseID) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	o.ability.background.run(func(ctx context.Context) {
		// KeepAlive renews the lease until the context is canceled; the
		// responses are drained so that the client never blocks on them.
		responses, keepAliveErr := client.KeepAlive(ctx, lease)
		if keepAliveErr != nil {
			return
		}

		for range responses {
		}
	})

	return nil
}
