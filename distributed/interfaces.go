package distributed

import (
	"context"

	"github.com/tursom/GoCollections/exceptions"
)

type (
	Router[Msg any] interface {
		// Send 向指定 id 的节点发送消息
		// jmp 消息已经经过了几个节点的转发
		Send(
			ctx context.Context, id string, msg Msg, jmp uint32,
		) (bool, exceptions.Exception)

		// Find 查询到目标节点的拓扑距离
		// return UNREACHABLE 在 e 为 nil 时表示节点不可达，否则表示发生错误
		Find(
			ctx context.Context, id string, jmp uint32,
		) (uint32, exceptions.Exception)
	}
)
