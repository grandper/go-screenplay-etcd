package etcd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/grandper/go-screenplay-etcd/etcd"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	t.Run("describe the mistake", func(t *testing.T) {
		t.Parallel()

		require.EqualError(t, etcd.ErrKeyNotProvided, "the key must be provided")
		require.EqualError(t, etcd.ErrPrefixNotProvided, "the prefix must be provided")
		require.EqualError(t, etcd.ErrNameNotProvided, "the name must be provided")
		require.EqualError(t, etcd.ErrEndpointNotProvided, "the endpoint must be provided")
		require.EqualError(t, etcd.ErrKeyNotFound, "was not found")
		require.EqualError(t, etcd.ErrKeyAlreadyExists, "already exists")
		require.EqualError(t, etcd.ErrKeyModified, "was modified")
		require.EqualError(t, etcd.ErrNoEndpointProvided, "no etcd endpoint was provided to the ability")
		require.EqualError(t, etcd.ErrFailedToConnect, "failed to connect to the etcd server")
		require.EqualError(t, etcd.ErrFailedToLoadTheCertificates, "failed to load the certificates")
		require.EqualError(t, etcd.ErrAlreadyWatching, "already watched")
		require.EqualError(t, etcd.ErrNotWatching, "not watched")
		require.EqualError(t, etcd.ErrLockAlreadyHeld, "already held by the actor")
		require.EqualError(t, etcd.ErrLockNotHeld, "not held by the actor")
		require.EqualError(t, etcd.ErrAlreadyCampaigning, "the actor already campaigns")
		require.EqualError(t, etcd.ErrNotCampaigning, "the actor does not campaign")
	})

	t.Run("read as a sentence once wrapped", func(t *testing.T) {
		t.Parallel()

		notFound := fmt.Errorf("the key %q %w", "does/not/exist", etcd.ErrKeyNotFound)
		require.ErrorIs(t, notFound, etcd.ErrKeyNotFound)
		require.EqualError(t, notFound, `the key "does/not/exist" was not found`)

		exists := fmt.Errorf("the key %q %w", "app/config/flag", etcd.ErrKeyAlreadyExists)
		require.ErrorIs(t, exists, etcd.ErrKeyAlreadyExists)
		require.EqualError(t, exists, `the key "app/config/flag" already exists`)

		modified := fmt.Errorf("the key %q %w since the revision %d", "app/config/counter", etcd.ErrKeyModified, 42)
		require.ErrorIs(t, modified, etcd.ErrKeyModified)
		require.EqualError(t, modified, `the key "app/config/counter" was modified since the revision 42`)

		notWatched := fmt.Errorf("the key %q is %w", "app/status", etcd.ErrNotWatching)
		require.ErrorIs(t, notWatched, etcd.ErrNotWatching)
		require.EqualError(t, notWatched, `the key "app/status" is not watched`)
	})

	t.Run("report a held lock with the error of the etcd client", func(t *testing.T) {
		t.Parallel()

		require.ErrorIs(t, etcd.ErrLocked, concurrency.ErrLocked)
	})

	t.Run("are distinct from one another", func(t *testing.T) {
		t.Parallel()

		errs := []error{
			etcd.ErrKeyNotProvided,
			etcd.ErrPrefixNotProvided,
			etcd.ErrNameNotProvided,
			etcd.ErrEndpointNotProvided,
			etcd.ErrKeyNotFound,
			etcd.ErrKeyAlreadyExists,
			etcd.ErrKeyModified,
			etcd.ErrNoEndpointProvided,
			etcd.ErrFailedToConnect,
			etcd.ErrFailedToLoadTheCertificates,
			etcd.ErrAlreadyWatching,
			etcd.ErrNotWatching,
			etcd.ErrLocked,
			etcd.ErrLockAlreadyHeld,
			etcd.ErrLockNotHeld,
			etcd.ErrAlreadyCampaigning,
			etcd.ErrNotCampaigning,
		}

		for i, err := range errs {
			for j, other := range errs {
				if i != j {
					require.NotErrorIs(t, err, other)
				}
			}
		}
	})
}
