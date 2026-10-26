package use

import impl "github.com/grandper/go-screenplay-etcd/internal/ability"

// ETCD enables the actor to use an etcd store; it reads "use etcd". The
// ability is configured fluently, reading "use etcd with ...":
//
//	use.ETCD().WithTheEndpoints("localhost:2379").WithTimingOutAfter(5).Seconds()
func ETCD() *impl.UseETCDAbility {
	return impl.UseETCD()
}
