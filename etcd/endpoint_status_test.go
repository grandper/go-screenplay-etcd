package etcd_test

import (
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay-etcd/etcd"
	"github.com/grandper/go-screenplay-etcd/fixture"
	"github.com/grandper/go-screenplay-etcd/question/alarms"
	questionfixture "github.com/grandper/go-screenplay-etcd/question/fixture"
	"github.com/grandper/go-screenplay-etcd/question/status"
	"github.com/grandper/go-screenplay-etcd/use"
)

func TestEndpointStatus(t *testing.T) {
	server := fixture.RunETCDServer(t)
	alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

	t.Run("exposes the parts of the status a member answered with", func(t *testing.T) {
		theStatus, err := alice.AsksFor(status.OfTheEndpoint(server.Endpoint()))

		require.NoError(t, err)
		assert.NotEmpty(t, theStatus.Version())
		assert.NotZero(t, theStatus.Leader(), "the only member of the cluster leads it")
		assert.Positive(t, theStatus.DBSize())
		assert.False(t, theStatus.IsLearner())
		assert.Empty(t, theStatus.Errors(), "a healthy member reports no error")
	})
}

func TestAlarm(t *testing.T) {
	t.Parallel()

	t.Run("exposes the parts of an alarm a member raised", func(t *testing.T) {
		t.Parallel()

		alarm := questionfixture.CreateFakeAlarm(42, "NOSPACE")

		assert.Equal(t, uint64(42), alarm.MemberID())
		assert.Equal(t, "NOSPACE", alarm.Type())
	})

	t.Run("is what the question about the alarms answers with", func(t *testing.T) {
		t.Parallel()
		server := fixture.RunETCDServer(t)
		alice := screenplay.ActorNamed("Alice").WhoCan(use.ETCD().WithTheEndpoints(server.Endpoint()))

		theAlarms, err := alice.AsksFor(alarms.OfTheCluster())

		require.NoError(t, err)
		assert.IsType(t, []*etcd.Alarm{}, theAlarms)
		assert.Empty(t, theAlarms, "a healthy cluster raises no alarm")
	})
}
