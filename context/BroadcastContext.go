package context

import (
	"sync"

	"gitea.tursom.cn/tursom/kvs/kv"
	log "github.com/sirupsen/logrus"
	"github.com/tursom/GoCollections/concurrent"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/util/bloom"
	"github.com/tursom/polycephalum"
	m2 "github.com/tursom/polycephalum/proto/m"

	_ "github.com/tursom/polycephalum"

	"github.com/tursom/tursom-im/conn"
	m "github.com/tursom/tursom-im/proto/msg"
)

type (
	// BroadcastContext
	// 负责实现广播的服务
	BroadcastContext interface {
		lang.Object
		Listen(channel *m2.BroadcastChannel, c conn.Conn) exceptions.Exception
		CancelListen(channel *m2.BroadcastChannel, conn conn.Conn) exceptions.Exception
		Send(channelType uint32, channel string, msg *m.ImMsg)
	}

	DistributeBroadcastContext interface {
		BroadcastContext
		AddNode(id string) exceptions.Exception
		SuspectNode(id string) exceptions.Exception
		UpdateFilter(id string, filter *bloom.Bloom) exceptions.Exception
		RemoteListen(id string, channel *m2.BroadcastChannel) exceptions.Exception
	}

	// localBroadcastContext
	// 纯本地广播实现
	localBroadcastContext struct {
		lang.BaseObject
		polycephalum    polycephalum.Polycephalum[*m.ImMsg]
		channelGroupMap map[*m2.BroadcastChannel]*conn.ConnGroup
		mutex           concurrent.RWLock
	}
)

func NewBroadcastContext() BroadcastContext {
	// TODO
	p := polycephalum.New[*m.ImMsg](
		"",
		kv.ProtoCodec(func() *m.ImMsg { return &m.ImMsg{} }),
		nil,
		nil, nil,
		nil,
	)

	return &localBroadcastContext{
		channelGroupMap: make(map[*m2.BroadcastChannel]*conn.ConnGroup),
		mutex:           new(sync.RWMutex),
		polycephalum:    p,
	}
}

func (b *localBroadcastContext) Listen(channel *m2.BroadcastChannel, c conn.Conn) exceptions.Exception {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	log.WithFields(log.Fields{
		"type":    channel.Type,
		"channel": channel.Channel,
	}).Info("listen local broadcast")

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

func (b *localBroadcastContext) CancelListen(channel *m2.BroadcastChannel, conn conn.Conn) exceptions.Exception {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	log.WithFields(log.Fields{
		"type":    channel.Type,
		"channel": channel.Channel,
	}).Info("cancel listen local broadcast")

	connGroup := b.channelGroupMap[channel]
	if connGroup == nil {
		return nil
	}
	connGroup.Remove(conn)
	if connGroup.Size() == 0 {
		delete(b.channelGroupMap, channel)
	}

	return nil
}

func (b *localBroadcastContext) Send(channelType uint32, channel string, msg *m.ImMsg) {
	b.polycephalum.Broadcast(channelType, channel, msg, nil)
}

func (b *localBroadcastContext) Receive(channel *m2.BroadcastChannel, msg *m.ImMsg) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	connGroup := b.channelGroupMap[channel]
	if connGroup != nil {
		connGroup.WriteChatMsg(msg, nil)
	}
	return
}
