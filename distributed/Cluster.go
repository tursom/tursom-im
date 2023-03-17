package distributed

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/lang/atomic"

	"github.com/tursom-im/exception"
)

const (
	UNREACHABLE = math.MaxUint32
)

type (
	// MessageProcessor 用于处理底层消息传输的接口
	// 由用户负责实现
	MessageProcessor[Msg any] interface {
		// LocalId 本地节点 ID
		LocalId() string

		// Send 对目标节点发送消息转发请求
		// return false 表示节点不可达
		Send(
			ctx context.Context, nextJmp string, target string, msg Msg, jmp uint32,
		) (bool, exceptions.Exception)

		// Find 对 nextJmp 节点发送路由查询请求
		//
		// return 从本节点出发，经 nextJmp 到 target 的拓扑距离。
		//        UNREACHABLE 表示节点不可达
		Find(
			ctx context.Context, nextJmp string, target string, jmp uint32,
		) (uint32, exceptions.Exception)
	}

	// Cluster 负责集群管理、消息转发等
	Cluster[Msg any] interface {
		// Send 向指定 id 的节点发送消息
		// jmp 消息已经经过了几个节点的转发
		Send(
			ctx context.Context, id string, msg Msg, jmp uint32,
		) (bool, exceptions.Exception)

		// Find 查询到目标节点的拓扑距离
		// return UNREACHABLE 表示节点不可达
		Find(
			ctx context.Context, id string, jmp uint32,
		) (uint32, exceptions.Exception)

		// AddNodes 添加集群节点记录
		AddNodes(nodes []string)

		// SetConnected 设置已连接的节点列表
		SetConnected(nodes []string)
	}

	clusterImpl[Msg any] struct {
		processor MessageProcessor[Msg]
		maxJmp    uint32
		connected map[string]*node
		nodes     map[string]*node
		lock      sync.Mutex
		cluster   atomic.Reference[clusterNodes]
	}

	clusterNodes struct {
		version uint32
		nodes   map[string]byte
	}

	node struct {
		id             string
		clusterVersion uint32
		nextJmp        nextJmpReference
		lock           sync.Mutex
	}

	nextJmpReference struct {
		atomic.Reference[nextJmpNode]
	}

	nextJmpNode struct {
		node     *node
		distance uint32
	}
)

var (
	unreachable = &nextJmpNode{distance: UNREACHABLE}
)

func NewCluster[Msg any](config *ClusterConfig[Msg]) Cluster[Msg] {
	c := &clusterImpl[Msg]{
		processor: config.processor,
		maxJmp:    config.maxJmp,
		connected: make(map[string]*node),
		nodes:     make(map[string]*node),
	}
	c.cluster.Store(&clusterNodes{})
	return c
}

func (c *clusterImpl[Msg]) Send(
	ctx context.Context, id string, msg Msg, jmp uint32,
) (bool, exceptions.Exception) {
	if jmp > c.maxJmp {
		return false, nil
	}

	node := c.nodeOf(id)

	topology := c.cluster.Load()
	if node.clusterVersion != topology.version {
		if _, ok := topology.nodes[id]; !ok {
			return false, exception.NewNodeOfflineException(fmt.Sprintf("node %s is offline", id))
		}
	}

	ok, e := c.sendToNode(ctx, node, msg, jmp+1)
	if !ok && e != nil {
		c.removeNode(id)
	}

	return ok, e
}

func (c *clusterImpl[Msg]) Find(ctx context.Context, id string, jmp uint32) (uint32, exceptions.Exception) {
	if c.processor.LocalId() == id {
		return 0, nil
	}

	if jmp > c.maxJmp {
		return UNREACHABLE, nil
	}

	if _, ok := c.cluster.Load().nodes[id]; !ok {
		return UNREACHABLE, nil
	}

	n := c.nodeOf(id)
	if !n.nextJmp.unreachable() {
		return n.nextJmp.distance() + jmp, nil
	}

	c.discovery(ctx, n, jmp)

	return n.nextJmp.distance(), nil
}

func (c *clusterImpl[Msg]) AddNodes(nodes []string) {
	topologySnapshot := c.cluster.Load()

	m, changed := sum(topologySnapshot.nodes, nodes)
	for changed && !c.cluster.CompareAndSwap(topologySnapshot, &clusterNodes{
		version: topologySnapshot.version + 1,
		nodes:   m,
	}) {
		m, changed = sum(topologySnapshot.nodes, nodes)
		topologySnapshot = c.cluster.Load()
	}
}

func (c *clusterImpl[Msg]) SetConnected(nodes []string) {
	c.AddNodes(nodes)

	connected := make(map[string]*node)
	for _, id := range nodes {
		connected[id] = c.nodeOf(id)
	}

	c.connected = connected
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

// removeNode 从集群中删除节点
func (c *clusterImpl[Msg]) removeNode(node string) {
	topologySnapshot := c.cluster.Load()
	if _, ok := topologySnapshot.nodes[node]; !ok {
		return
	}

	m := make(map[string]byte)
	for s := range topologySnapshot.nodes {
		m[s] = 0
	}
	delete(m, node)

	for !c.cluster.CompareAndSwap(topologySnapshot, &clusterNodes{
		version: topologySnapshot.version + 1,
		nodes:   m,
	}) {
		topologySnapshot = c.cluster.Load()
		if _, ok := topologySnapshot.nodes[node]; !ok {
			return
		}

		m = make(map[string]byte)
		for s := range topologySnapshot.nodes {
			m[s] = 0
		}
		delete(m, node)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.nodes, node)
	delete(c.connected, node)
}

func (c *clusterImpl[Msg]) sendToNode(ctx context.Context, node *node, msg Msg, jmp uint32) (bool, exceptions.Exception) {
	if node.nextJmp.unreachable() {
		c.discovery(ctx, node, jmp)

		if node.nextJmp.unreachable() {
			return false, exception.NewNodeOfflineException(fmt.Sprintf("node %s is offline", node.id))
		}
	}

	return c.send(ctx, node, msg, jmp)
}

func (c *clusterImpl[Msg]) send(ctx context.Context, node *node, msg Msg, jmp uint32) (bool, exceptions.Exception) {
	if ok, e := c.processor.Send(ctx, node.nextJmp.nodeId(), node.id, msg, jmp); ok || e != nil {
		return ok, e
	}

	// 发送失败，尝试重新查找节点并发送消息
	// 最多重试 3 次
	for retryTimes := 0; retryTimes < 3; retryTimes++ {
		log.Warnf("failed to send msg to %s, retry %d times", node.id, retryTimes+1)

		if _, ok := c.nodes[node.id]; !ok {
			return false, nil
		}

		node.nextJmp.Store(unreachable)

		c.discovery(ctx, node, jmp)

		// 不可达，重新尝试
		if node.nextJmp.unreachable() {
			continue
		}

		// 如果发送失败则重试
		if send, e := c.processor.Send(ctx, node.nextJmp.nodeId(), node.id, msg, jmp); send || e != nil {
			return send, e
		}
	}

	return false, nil
}

// discovery 尝试发现目标节点
// jmp 查询请求已经经过的节点数量
func (c *clusterImpl[Msg]) discovery(ctx context.Context, n *node, jmp uint32) {
	// 本机节点直接返回
	if n.id == c.processor.LocalId() {
		n.nextJmp.Store(&nextJmpNode{node: n, distance: 0})
		return
	}

	// 上锁，防止同时进行冗余查询，浪费系统资源
	n.lock.Lock()
	defer n.lock.Unlock()

	// double check，确认是否已经找到目标节点
	if !n.nextJmp.unreachable() {
		return
	}

	if _, ok := ctx.Deadline(); !ok {
		//goland:noinspection GoVetLostCancel
		ctx, _ = context.WithTimeout(ctx, 45*time.Second)
	}

	// connected nodes snapshot
	connected := c.connected

	finishedCount := new(atomic.Int32)
	done := make(lang.RawChannel[bool])

	// 遍历已连接节点，发送寻人启事
	for _, next := range connected {
		njn := &nextJmpNode{
			node:     next,
			distance: 1,
		}

		// 发现下一跳正好是目标节点，立即存储
		// 但是由于分布式环境充满不确定性，还是要发送确认消息的
		if next.id == n.id {
			n.nextJmp.Store(njn)
		}

		// 启动新 goroutine 发送请求
		go c.check(ctx, next, n, jmp, func(node string, distance uint32, e exceptions.Exception) {
			if distance == UNREACHABLE || e != nil {
				n.nextJmp.CompareAndSwap(njn, unreachable)

				// 确保在所有节点都返回不可达时，能够正确返回
				if finishedCount.Add(1) == int32(len(connected)) {
					done.TrySend(true)
				}

				if e != nil {
					log.Warnf("failed to check for node %s: %s", next.id, e)
					e.PrintStackTrace()
				}
			}

			n.nextJmp.update(&nextJmpNode{
				node:     next,
				distance: distance,
			})

			done.TrySend(true)
		})
	}

	if _, ok := done.ReceiveTimeout(time.Minute); !ok {
		log.Warnf("discovery to %s timeout", n.id)
	}
}

func (c *clusterImpl[Msg]) check(
	ctx context.Context,
	nextJmp, n *node,
	jmp uint32,
	done func(node string, distance uint32, e exceptions.Exception),
) {
	distance, e := c.processor.Find(ctx, nextJmp.id, n.id, jmp)
	if e != nil {
		done(n.id, UNREACHABLE, e)
		return
	}

	done(n.id, distance, nil)
}

// nodeOf 根据 id 获取 node 实例
// node 作为 cluster 中的单例实现
func (c *clusterImpl[Msg]) nodeOf(id string) *node {
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

	n.nextJmp.Store(unreachable)

	return n
}

func (n *nextJmpReference) unreachable() bool {
	nextJmp := n.Load()
	return nextJmp == nil || nextJmp.distance == UNREACHABLE
}

func (n *nextJmpReference) nodeId() string {
	return n.Load().node.id
}

func (n *nextJmpReference) distance() uint32 {
	return n.Load().distance
}

func (n *nextJmpReference) update(nextJmp *nextJmpNode) {
	if nextJmp.distance == UNREACHABLE {
		return
	}

	snap := n.Load()
	for (snap == nil || snap.distance > nextJmp.distance) &&
		!n.CompareAndSwap(snap, nextJmp) {

		snap = n.Load()
	}
}
