package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/lang/atomic"

	"github.com/tursom-im/exception"
)

type (
	MessageProcessor[Msg any] interface {
		LocalId() string

		Send(
			ctx context.Context, nextJmp string, target string, msg Msg, jmp uint32,
		) (bool, exceptions.Exception)

		Find(
			ctx context.Context, nextJmp string, target string, jmp uint32,
		) (int32, exceptions.Exception)
	}

	Cluster[Msg any] interface {
		Send(
			ctx context.Context, id string, msg Msg, jmp uint32,
		) (bool, exceptions.Exception)

		Find(
			ctx context.Context, id string, distance int32, jmp uint32,
		) (int32, exceptions.Exception)

		AddNodes(nodes []string)
		ConnectedNodes(nodes []string)
	}

	clusterImpl[Msg any] struct {
		processor MessageProcessor[Msg]
		maxJmp    uint32
		version   uint32
		connected map[string]*node
		nodes     map[string]*node
		lock      sync.Mutex
		topology  atomic.Reference[topology]
	}

	topology struct {
		version uint32
		nodes   map[string]byte
	}

	node struct {
		id              string
		snapshotVersion uint32
		nextJmp         *node
		distance        uint32
		lock            sync.Mutex
	}

	nextJmpNode struct {
		node     *node
		distance uint32
	}
)

func NewCluster[Msg any](config *ClusterConfig[Msg]) Cluster[Msg] {
	c := &clusterImpl[Msg]{
		processor: config.processor,
		maxJmp:    config.maxJmp,
		connected: make(map[string]*node),
		nodes:     make(map[string]*node),
	}
	c.topology.Store(&topology{})
	return c
}

func (c *clusterImpl[Msg]) Send(
	ctx context.Context, id string, msg Msg, jmp uint32,
) (bool, exceptions.Exception) {
	if jmp > c.maxJmp {
		return false, nil
	}

	node := c.getNode(id)

	topology := c.topology.Load()
	if node.snapshotVersion != topology.version {
		if _, ok := topology.nodes[id]; !ok {
			return false, exception.NewNodeOfflineException(fmt.Sprintf("node %s is offline", id))
		}
	}

	return c.sendToNode(ctx, node, msg, jmp+1)
}

func (c *clusterImpl[Msg]) Find(ctx context.Context, id string, distance int32, jmp uint32) (int32, exceptions.Exception) {
	if jmp > c.maxJmp {
		return -1, nil
	}

	if c.processor.LocalId() == id {
		return 0, nil
	}

	_, ok := c.topology.Load().nodes[id]
	if !ok {
		return -1, nil
	}
	n := c.getNode(id)
	if n.distance > 0 {
		return int32(n.distance) + distance, nil
	}

	if e := c.discovery(ctx, n, jmp); e != nil {
		return -1, e
	}

	return int32(n.distance), nil
}

func (c *clusterImpl[Msg]) AddNodes(nodes []string) {
	topologySnapshot := c.topology.Load()

	m, changed := sum(topologySnapshot.nodes, nodes)
	for changed && !c.topology.CompareAndSwap(topologySnapshot, &topology{
		version: topologySnapshot.version + 1,
		nodes:   m,
	}) {
		m, changed = sum(topologySnapshot.nodes, nodes)
		topologySnapshot = c.topology.Load()
	}
}

func (c *clusterImpl[Msg]) ConnectedNodes(nodes []string) {
	c.AddNodes(nodes)

	c.lock.Lock()
	defer c.lock.Unlock()

	c.connected = make(map[string]*node)
	for _, id := range nodes {
		n := c.nodes[id]
		if n == nil {
			n = &node{id: id}
			c.nodes[id] = n
		}
		c.connected[id] = n
	}
}

func sum(m map[string]byte, s []string) (map[string]byte, bool) {
	newMap := make(map[string]byte)

	for id := range m {
		newMap[id] = 0
	}

	chanced := false
	for _, id := range s {
		if _, ok := newMap[id]; !ok {
			chanced = true
		}
		newMap[id] = 0
	}

	return newMap, chanced
}

func (c *clusterImpl[Msg]) removeNode(node string) {
	topologySnapshot := c.topology.Load()
	if _, ok := topologySnapshot.nodes[node]; !ok {
		return
	}

	m := make(map[string]byte)
	for s := range topologySnapshot.nodes {
		m[s] = 0
	}
	delete(m, node)

	for !c.topology.CompareAndSwap(topologySnapshot, &topology{
		version: topologySnapshot.version + 1,
		nodes:   m,
	}) {
		topologySnapshot = c.topology.Load()
		if _, ok := topologySnapshot.nodes[node]; !ok {
			return
		}

		m = make(map[string]byte)
		for s := range topologySnapshot.nodes {
			m[s] = 0
		}
		delete(m, node)
	}
}

func (c *clusterImpl[Msg]) sendToNode(ctx context.Context, node *node, msg Msg, jmp uint32) (bool, exceptions.Exception) {
	if node.nextJmp == nil {
		if e := c.discovery(ctx, node, jmp); e != nil {
			return false, e
		}

		if node.nextJmp == nil {
			return false, exception.NewNodeOfflineException(fmt.Sprintf("node %s is offline", node.id))
		}
	}

	return c.send(ctx, node, msg, jmp)
}

func (c *clusterImpl[Msg]) send(ctx context.Context, node *node, msg Msg, jmp uint32) (bool, exceptions.Exception) {
	ok, e := c.processor.Send(ctx, node.nextJmp.id, node.id, msg, jmp)
	if ok || e != nil {
		return ok, e
	}

	if _, ok := c.nodes[node.id]; !ok {
		return false, nil
	}

	node.nextJmp = nil
	if e := c.discovery(ctx, node, jmp); e != nil {
		return false, e
	}

	if node.nextJmp == nil {
		return false, nil
	}

	return c.processor.Send(ctx, node.nextJmp.id, node.id, msg, jmp)
}

func (c *clusterImpl[Msg]) discovery(ctx context.Context, n *node, jmp uint32) exceptions.Exception {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.nextJmp != nil {
		return nil
	}

	if n.id == c.processor.LocalId() {
		n.nextJmp = n
		return nil
	}

	_, ok := ctx.Deadline()
	if !ok {
		ctx, _ = context.WithTimeout(ctx, 45*time.Second)
	}

	connected := c.connected

	finishedCount := new(atomic.Int32)
	var done lang.Channel[bool] = make(lang.RawChannel[bool])

	var nextJmp *nextJmpNode

	for _, next := range connected {
		go c.check(ctx, next, n, jmp, func(node string, distance int32, e exceptions.Exception) {
			if distance <= 0 {
				if finishedCount.Add(1) == int32(len(connected)) {
					done.TrySend(true)
				}
				return
			}

			nextJmp = &nextJmpNode{
				node:     next,
				distance: uint32(distance),
			}

			go func() {
				n.lock.Lock()
				defer n.lock.Unlock()

				if n.distance > uint32(distance) {
					n.nextJmp = next
					n.distance = uint32(distance)
				}
			}()

			done.TrySend(true)
		})
	}

	_, ok = done.ReceiveTimeout(time.Minute)
	if !ok {
		return exception.NewTimeoutException("")
	}

	nextJmp0 := nextJmp
	if nextJmp0 != nil {
		n.nextJmp = nextJmp0.node
		n.distance = nextJmp0.distance
	}
	return nil

}

func (c *clusterImpl[Msg]) check(ctx context.Context, nextJmp, n *node, jmp uint32, done func(node string, distance int32, e exceptions.Exception)) {
	distance, e := c.processor.Find(ctx, nextJmp.id, n.id, jmp)
	if e != nil {
		done(n.id, -1, e)
		return
	}

	done(n.id, distance, nil)

	if distance > 0 {

		n.lock.Lock()
		defer n.lock.Unlock()

		if n.distance > uint32(distance) {
			n.nextJmp = nextJmp
			n.distance = uint32(distance)
		}
	}
}

func (c *clusterImpl[Msg]) getNode(id string) *node {
	n := c.nodes[id]
	if n != nil {
		return n
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// double check
	n = c.nodes[id]
	if n != nil {
		return n
	}

	n = &node{id: id}
	c.nodes[id] = n

	return n
}
