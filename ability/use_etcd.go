package ability

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

// UseETCD enables the actor to use an etcd store. The ability is configured
// fluently, reading "use etcd with ...":
//
//	UseETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(5).Seconds()
func UseETCD() *impl.UseETCDAbility {
	return impl.UseETCD()
}
