package context

import (
	"sync"

	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/conn"
	m "github.com/tursom-im/proto/msg"
)

type (
	// BroadcastContext
	// 负责实现广播的服务
	BroadcastContext interface {
		lang.Object
		Listen(channel int32, c conn.Conn) exceptions.Exception
		CancelListen(channel int32, conn conn.Conn) exceptions.Exception
		Send(channel int32, msg *m.ImMsg, filter func(conn.Conn) bool) int
	}

	// localBroadcastContext
	// 纯本地广播实现
	localBroadcastContext struct {
		lang.BaseObject
		channelGroupMap map[int32]*conn.ConnGroup
		mutex           concurrent.RWLock
	}
)

func NewBroadcastContext() BroadcastContext {
	return &localBroadcastContext{
		channelGroupMap: make(map[int32]*conn.ConnGroup),
		mutex:           new(sync.RWMutex),
	}
}

func (b *localBroadcastContext) Listen(channel int32, c conn.Conn) exceptions.Exception {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup == nil {
		connGroup = conn.NewConnGroup()
		b.channelGroupMap[channel] = connGroup
	}
	connGroup.Add(c)

	c.AddEventListener(func(event conn.Event) {
		if !event.EventId().IsConnClosed() || connGroup.Size() != 0 {
			return
		}
		b.mutex.Lock()
		defer b.mutex.Unlock()

		delete(b.channelGroupMap, channel)
	})

	return nil
}

func (b *localBroadcastContext) CancelListen(channel int32, conn conn.Conn) exceptions.Exception {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup == nil {
		return nil
	}
	connGroup.Remove(conn)

	return nil
}

func (b *localBroadcastContext) Send(channel int32, msg *m.ImMsg, filter func(conn.Conn) bool) int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup != nil {
		return connGroup.WriteChatMsg(msg, filter)
	}
	return 0
}
