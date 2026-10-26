package ability_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/grandper/go-screenplay-etcd/internal/ability"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	t.Run("describe the mistake", func(t *testing.T) {
		t.Parallel()

		require.EqualError(t, ability.ErrNoEndpointProvided, "no etcd endpoint was provided to the ability")
		require.EqualError(t, ability.ErrFailedToConnect, "failed to connect to the etcd server")
		require.EqualError(t, ability.ErrFailedToLoadTheCertificates, "failed to load the certificates")
		require.EqualError(t, ability.ErrAlreadyWatching, "already watched")
		require.EqualError(t, ability.ErrNotWatching, "not watched")
		require.EqualError(t, ability.ErrLockAlreadyHeld, "already held by the actor")
		require.EqualError(t, ability.ErrLockNotHeld, "not held by the actor")
		require.EqualError(t, ability.ErrAlreadyCampaigning, "the actor already campaigns")
		require.EqualError(t, ability.ErrNotCampaigning, "the actor does not campaign")
	})

	t.Run("read as a sentence once wrapped", func(t *testing.T) {
		t.Parallel()

		held := fmt.Errorf("the lock %q is %w", "deployment", ability.ErrLockAlreadyHeld)
		require.ErrorIs(t, held, ability.ErrLockAlreadyHeld)
		require.EqualError(t, held, `the lock "deployment" is already held by the actor`)

		campaigning := fmt.Errorf("%w in the election %q", ability.ErrAlreadyCampaigning, "primary-worker")
		require.ErrorIs(t, campaigning, ability.ErrAlreadyCampaigning)
		require.EqualError(t, campaigning, `the actor already campaigns in the election "primary-worker"`)
	})

	t.Run("report a held lock with the error of the etcd client", func(t *testing.T) {
		t.Parallel()

		require.ErrorIs(t, ability.ErrLocked, concurrency.ErrLocked)
	})

	t.Run("are distinct from one another", func(t *testing.T) {
		t.Parallel()

		errs := []error{
			ability.ErrNoEndpointProvided,
			ability.ErrFailedToConnect,
			ability.ErrFailedToLoadTheCertificates,
			ability.ErrAlreadyWatching,
			ability.ErrNotWatching,
			ability.ErrLocked,
			ability.ErrLockAlreadyHeld,
			ability.ErrLockNotHeld,
			ability.ErrAlreadyCampaigning,
			ability.ErrNotCampaigning,
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
