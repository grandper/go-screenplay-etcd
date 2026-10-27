package ability

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"

	"github.com/grandper/go-screenplay-etcd/internal/answer"
)

// ToPut puts the value for the key, creating the key or overwriting its value.
// The key is attached to the lease, so that it is removed when the lease
// expires; the lease zero attaches the key to no lease.
func (o *Operations) ToPut(ctx context.Context, key, value string, lease clientv3.LeaseID) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	_, err = client.Put(ctx, key, value, withTheLease(lease)...)

	return err
}

// withTheLease returns the option attaching a key to the lease, or no option
// for the lease zero.
func withTheLease(lease clientv3.LeaseID) []clientv3.OpOption {
	if lease == clientv3.NoLease {
		return nil
	}

	return []clientv3.OpOption{clientv3.WithLease(lease)}
}

// ToRetrieveTheKey retrieves the key with its value and its metadata. A key
// that does not exist is answered as a Key that does not exist, not as an
// error.
func (o *Operations) ToRetrieveTheKey(ctx context.Context, key string) (*answer.Key, error) {
	client, err := o.client()
	if err != nil {
		return nil, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(response.Kvs) == 0 {
		return answer.NewMissingKey(key), nil
	}

	return answer.NewKey(response.Kvs[0]), nil
}

// ToRemove removes the key. Removing a key that does not exist is not an error.
func (o *Operations) ToRemove(ctx context.Context, key string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	_, err = client.Delete(ctx, key)

	return err
}

// ToRemoveTheKeys removes the keys in one transaction, so that the store never
// holds some of them without the others. Removing a key that does not exist is
// not an error.
func (o *Operations) ToRemoveTheKeys(ctx context.Context, keys ...string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	operations := make([]clientv3.Op, 0, len(keys))
	for _, key := range keys {
		operations = append(operations, clientv3.OpDelete(key))
	}

	_, err = client.Txn(ctx).Then(operations...).Commit()

	return err
}

// ToRemoveTheKeysStartingWith removes every key starting with the prefix.
// Removing a prefix no key starts with is not an error.
func (o *Operations) ToRemoveTheKeysStartingWith(ctx context.Context, prefix string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	_, err = client.Delete(ctx, prefix, clientv3.WithPrefix())

	return err
}

// ToRetrieveTheKeysStartingWith retrieves every key starting with the prefix,
// with its value and its metadata, in the order of the keys. When no key
// starts with the prefix, it answers an empty slice, not an error.
func (o *Operations) ToRetrieveTheKeysStartingWith(ctx context.Context, prefix string) ([]*answer.Key, error) {
	client, err := o.client()
	if err != nil {
		return nil, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	keys := make([]*answer.Key, 0, len(response.Kvs))
	for _, kv := range response.Kvs {
		keys = append(keys, answer.NewKey(kv))
	}

	return keys, nil
}

// ToPutIfAbsent puts the value for the key only when the key does not exist,
// in one transaction, attaching the key to the lease as ToPut does. It returns
// false, without an error, when the key exists.
func (o *Operations) ToPutIfAbsent(ctx context.Context, key, value string, lease clientv3.LeaseID) (bool, error) {
	return o.commit(ctx, clientv3util.KeyMissing(key), clientv3.OpPut(key, value, withTheLease(lease)...))
}

// ToPutIfUnchanged puts the value for the key only when the key was not
// modified since the revision, in one transaction, attaching the key to the
// lease as ToPut does. The revision zero means that the key must not exist.
// It returns false, without an error, when the key was modified.
func (o *Operations) ToPutIfUnchanged(
	ctx context.Context,
	key, value string,
	revision int64,
	lease clientv3.LeaseID,
) (bool, error) {
	return o.commit(
		ctx,
		clientv3.Compare(clientv3.ModRevision(key), "=", revision),
		clientv3.OpPut(key, value, withTheLease(lease)...),
	)
}

// ToRemoveIfUnchanged removes the key only when it was not modified since the
// revision, in one transaction. It returns false, without an error, when the
// key was modified.
func (o *Operations) ToRemoveIfUnchanged(ctx context.Context, key string, revision int64) (bool, error) {
	return o.commit(ctx, clientv3.Compare(clientv3.ModRevision(key), "=", revision), clientv3.OpDelete(key))
}

// ToPutTheValues puts every value for its key in one transaction, so that the
// store never holds some of them without the others.
func (o *Operations) ToPutTheValues(ctx context.Context, values map[string]string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	operations := make([]clientv3.Op, 0, len(values))
	for key, value := range values {
		operations = append(operations, clientv3.OpPut(key, value))
	}

	_, err = client.Txn(ctx).Then(operations...).Commit()

	return err
}

// commit runs the operation in a transaction guarded by the comparison, and
// returns whether the comparison held and the operation was applied.
func (o *Operations) commit(ctx context.Context, comparison clientv3.Cmp, operation clientv3.Op) (bool, error) {
	client, err := o.client()
	if err != nil {
		return false, err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	response, err := client.Txn(ctx).If(comparison).Then(operation).Commit()
	if err != nil {
		return false, err
	}

	return response.Succeeded, nil
}
