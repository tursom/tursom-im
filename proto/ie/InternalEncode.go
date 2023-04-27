package ie

import (
	"gitea.tursom.cn/tursom/kvs/kv"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type (
	sysKeyCodec struct {
		lang.BaseObject
	}
)

var (
	SysKeyCodec kv.Codec[*KVStoreKey, string] = &sysKeyCodec{}
)

func SysKey(key string) *KVStoreKey {
	return &KVStoreKey{
		Content: &KVStoreKey_System{
			System: key,
		},
	}
}

func (s *sysKeyCodec) Encode(v2 string) *KVStoreKey {
	return SysKey(v2)
}

func (s *sysKeyCodec) Decode(v1 *KVStoreKey) string {
	key, ok := v1.Content.(*KVStoreKey_System)
	if !ok {
		panic(exceptions.NewIllegalAccessException("illegal kvs key", exceptions.Cfg().SetCause(v1)))
	}

	return key.System
}
