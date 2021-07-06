package context

import (
	"github.com/tursom/GoCollections/exceptions"
	"sync"
	"tursom-im/im_conn"
)

// BroadcastContext
// 负责实现广播的服务
type BroadcastContext struct {
	channelGroupMap map[int32]*im_conn.ConnGroup
	mutex           *sync.RWMutex
}

func NewBroadcastContext() *BroadcastContext {
	return &BroadcastContext{
		channelGroupMap: make(map[int32]*im_conn.ConnGroup),
		mutex:           new(sync.RWMutex),
	}
}

func (b *BroadcastContext) Listen(channel int32, conn *im_conn.AttachmentConn) exceptions.Exception {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup == nil {
		connGroup = im_conn.NewConnGroup()
		b.channelGroupMap[channel] = connGroup
	}
	connGroup.Add(conn)

	conn.AddEventListener(func(event im_conn.ConnEvent) {
		if event.EventId().IsConnClosed() {
			if connGroup.Size() == 0 {
				b.mutex.Lock()
				defer b.mutex.Unlock()

				if connGroup.Size() == 0 {
					delete(b.channelGroupMap, channel)
				}
			}
		}
	})

	return nil
}

func (b *BroadcastContext) CancelListen(channel int32, conn *im_conn.AttachmentConn) exceptions.Exception {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup == nil {
		return nil
	}
	connGroup.Remove(conn)

	return nil
}

func (b *BroadcastContext) Send(channel int32, data []byte, filter func(*im_conn.AttachmentConn) bool) int32 {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup != nil {
		return connGroup.WriteBinaryFrame(data, filter)
	}
	return 0
}
