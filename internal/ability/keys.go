package ability

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ToPut puts the value for the key, creating the key or overwriting its value.
func (o *Operations) ToPut(ctx context.Context, key, value string) error {
	client, err := o.client()
	if err != nil {
		return err
	}

	ctx, cancel := o.ability.context(ctx)
	defer cancel()

	_, err = client.Put(ctx, key, value)

	return err
}

// ToRetrieveTheKey retrieves the key with its value and its metadata. A key
// that does not exist is answered as a Key that does not exist, not as an
// error.
func (o *Operations) ToRetrieveTheKey(ctx context.Context, key string) (*Key, error) {
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
		return NewMissingKey(key), nil
	}

	return NewKey(response.Kvs[0]), nil
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
