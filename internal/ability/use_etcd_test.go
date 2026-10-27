package ability_test

import (
	"sync"
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/internal/ability"
	"github.com/grandper/go-screenplay-etcd/internal/fixture"
)

func TestUseETCDAbility(t *testing.T) {
	server := fixture.RunETCDServer(t)

	t.Run("is safe for concurrent use", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(
			screenplay.ActorNamed("Alice").WhoCan(ability.UseETCD().WithTheEndpoints(server.Endpoint())),
		)

		const branches = 10
		errs := make(chan error, 2*branches)
		var group sync.WaitGroup
		for i := range branches {
			group.Go(func() {
				errs <- theAbility.ToPut(t.Context(), "concurrent/key", string(rune('a'+i)), 0)
				_, err := theAbility.ToRetrieveTheKey(t.Context(), "concurrent/key")
				errs <- err
			})
		}
		group.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}
	})

	t.Run("fails every operation when the actor is nil", func(t *testing.T) {
		theAbility := ability.TheAbilityOf(nil)

		require.ErrorIs(t, theAbility.ToPut(t.Context(), "k", "v", 0), screenplay.ErrActorMissingAbility)
	})

	t.Run("reports the ability that was not given on a nil receiver", func(t *testing.T) {
		var theAbility *ability.UseETCDAbility

		require.ErrorIs(t, theAbility.Forget(), screenplay.ErrActorMissingAbility)
		assert.Nil(t, theAbility.WithTheEndpoints("localhost:2379"))
		assert.Nil(t, theAbility.WithTheNamespace("ns/"))
		assert.Nil(t, theAbility.WithTheClient(nil))
		assert.Nil(t, theAbility.WithTheCertificateAuthority("ca.crt"))
		assert.Nil(t, theAbility.WithTheClientCertificate("client.crt").AndTheKey("client.key"))
	})

	t.Run("can be forgotten more than once", func(t *testing.T) {
		theAbility := ability.UseETCD().WithTheEndpoints(server.Endpoint())
		require.NoError(
			t,
			ability.TheAbilityOf(screenplay.ActorNamed("Alice").WhoCan(theAbility)).ToPut(t.Context(), "k", "v", 0),
		)

		require.NoError(t, theAbility.Forget())
		require.NoError(t, theAbility.Forget())
	})

	t.Run("implements the stringer interface", func(t *testing.T) {
		assert.Equal(t, "use etcd", ability.UseETCD().String())
	})
}
