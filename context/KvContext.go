package context

import (
	"gitea.tursom.cn/tursom/kvs/kv"

	"github.com/tursom/tursom-im/proto/ie"
)

type (
	KVContext interface {
		kv.Store[*ie.KVStoreKey, []byte]
	}
)

func SysKVC[V any](kvs kv.Store[*ie.KVStoreKey, V]) kv.Store[string, V] {
	return kv.KCodecStore(kvs, ie.SysKeyCodec)
}
