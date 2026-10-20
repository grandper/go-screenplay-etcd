package fixture

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/grandper/go-screenplay-etcd/ability"
	impl "github.com/grandper/go-screenplay-etcd/internal/ability"
)

// CreateFakeExistingKey creates a fake key that exists, holding the value, as
// if it had just been created at the first revision of the store.
func CreateFakeExistingKey(name, value string) *ability.Key {
	return impl.NewKey(&mvccpb.KeyValue{
		Key:            []byte(name),
		Value:          []byte(value),
		Version:        1,
		CreateRevision: 1,
		ModRevision:    1,
	})
}

// CreateFakeMissingKey creates a fake key that does not exist.
func CreateFakeMissingKey(name string) *ability.Key {
	return impl.NewMissingKey(name)
}

// CreateFakePutEvent creates a fake event that created the key with the value,
// at the first revision of the store.
func CreateFakePutEvent(key, value string) *ability.Event {
	return impl.NewEvent(&clientv3.Event{
		Type: mvccpb.PUT,
		Kv: &mvccpb.KeyValue{
			Key:            []byte(key),
			Value:          []byte(value),
			Version:        1,
			CreateRevision: 1,
			ModRevision:    1,
		},
	}, false)
}

// CreateFakeDeleteEvent creates a fake event that removed the key, which held
// the previous value, at the second revision of the store.
func CreateFakeDeleteEvent(key, previousValue string) *ability.Event {
	return impl.NewEvent(deleteEvent(key, previousValue, 0), false)
}

// CreateFakeExpiryEvent creates a fake event that removed the key, which held
// the previous value, because its lease expired.
func CreateFakeExpiryEvent(key, previousValue string) *ability.Event {
	return impl.NewEvent(deleteEvent(key, previousValue, 1), true)
}

// The revisions of the store at which a fake key is created and removed.
const (
	createdAtRevision = 1
	removedAtRevision = 2
)

// deleteEvent builds the event of the store removing the key, which held the
// previous value under the lease.
func deleteEvent(key, previousValue string, lease int64) *clientv3.Event {
	return &clientv3.Event{
		Type: mvccpb.DELETE,
		Kv: &mvccpb.KeyValue{
			Key:         []byte(key),
			ModRevision: removedAtRevision,
		},
		PrevKv: &mvccpb.KeyValue{
			Key:            []byte(key),
			Value:          []byte(previousValue),
			Version:        1,
			CreateRevision: createdAtRevision,
			ModRevision:    createdAtRevision,
			Lease:          lease,
		},
	}
}

// CreateFakeAlarm creates a fake alarm of the given type, such as "NOSPACE",
// raised by the member.
func CreateFakeAlarm(memberID uint64, alarmType string) *ability.Alarm {
	return impl.NewAlarm(memberID, alarmType)
}
