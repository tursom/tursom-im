package router

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

	"github.com/tursom-im/distributed"
	"github.com/tursom-im/exception"
)

const (
	UNREACHABLE = math.MaxUint32
)

type (
	// Processor 用于处理底层消息传输的接口
	// 由用户负责实现
	Processor[Msg any] interface {
		// LocalId 本地节点 ID
		LocalId() string

		// Send 对目标主机发送消息转发请求
		// return false, nil 表示节点不可达
		Send(
			ctx context.Context, nextJmp string, target string, msg Msg, jmp uint32,
		) (bool, exceptions.Exception)

		// Find 对 nextJmp 所代表的的主机发送路由查询请求
		//
		// return 从本机出发，经 nextJmp 到 target 的拓扑距离。
		//        UNREACHABLE 在 e 为 nil 时表示主机不可达，否则表示发生错误
		Find(
			ctx context.Context, nextJmp string, target string, jmp uint32,
		) (uint32, exceptions.Exception)
	}

	// Router 负责实现动态路由发现
	// 这只是个路由实现，网络主机管理需要用户负责
	Router[Msg any] interface {
		distributed.Router[Msg]

		// AddHosts 添加网络主机在线记录
		// 标记主机离线由 Router 本身负责
		AddHosts(hosts []string)

		// SetDirectly 设置可直达的主机列表
		SetDirectly(hosts []string)
	}

	routerImpl[Msg any] struct {
		processor Processor[Msg]
		maxJmp    uint32
		direct    map[string]*routeHost
		hosts     map[string]*routeHost
		lock      sync.Mutex
		network   atomic.Reference[network]
	}

	network struct {
		version uint32
		hosts   map[string]byte
	}

	routeHost struct {
		id             string
		networkVersion uint32
		nextJmp        nextJmpReference
		lock           sync.Mutex
	}

	nextJmpReference struct {
		atomic.Reference[nextJmpHost]
	}

	nextJmpHost struct {
		host     *routeHost
		distance uint32
	}
)

var (
	unreachable = &nextJmpHost{distance: UNREACHABLE}
)

func NewRouter[Msg any](config *Config[Msg]) Router[Msg] {
	c := &routerImpl[Msg]{
		processor: config.processor,
		maxJmp:    config.maxJmp,
		direct:    make(map[string]*routeHost),
		hosts:     make(map[string]*routeHost),
	}
	c.network.Store(&network{})
	return c
}

func (c *routerImpl[Msg]) Send(
	ctx context.Context, id string, msg Msg, jmp uint32,
) (bool, exceptions.Exception) {
	if jmp > c.maxJmp {
		return false, nil
	}

	host := c.hostOf(id)

	topology := c.network.Load()
	if host.networkVersion != topology.version {
		if _, ok := topology.hosts[id]; !ok {
			return false, exception.NewNodeOfflineException(fmt.Sprintf("routeHost %s is offline", id))
		}
	}

	ok, e := c.send(ctx, host, msg, jmp+1)
	if !ok && e == nil {
		c.removeNode(id)
	}

	return ok, e
}

func (c *routerImpl[Msg]) send(ctx context.Context, node *routeHost, msg Msg, jmp uint32) (bool, exceptions.Exception) {
	if node.nextJmp.unreachable() {
		c.discovery(ctx, node, jmp)

		if node.nextJmp.unreachable() {
			return false, nil
		}
	}

	if ok, e := c.processor.Send(ctx, node.nextJmp.hostId(), node.id, msg, jmp); ok || e != nil {
		return ok, e
	}

	// 发送失败，尝试重新查找节点并发送消息
	// 最多重试 3 次
	for retryTimes := 0; retryTimes < 3; retryTimes++ {
		log.Warnf("failed to send0 msg to %s, retry %d times", node.id, retryTimes+1)

		if _, ok := c.hosts[node.id]; !ok {
			return false, nil
		}

		node.nextJmp.Store(unreachable)

		c.discovery(ctx, node, jmp)

		// 不可达，重新尝试
		if node.nextJmp.unreachable() {
			continue
		}

		// 如果发送失败则重试
		if send, e := c.processor.Send(ctx, node.nextJmp.hostId(), node.id, msg, jmp); send || e != nil {
			return send, e
		}
	}

	return false, nil
}

func (c *routerImpl[Msg]) Find(ctx context.Context, id string, jmp uint32) (uint32, exceptions.Exception) {
	if c.processor.LocalId() == id {
		return 0, nil
	}

	if jmp > c.maxJmp {
		return UNREACHABLE, nil
	}

	if _, ok := c.network.Load().hosts[id]; !ok {
		return UNREACHABLE, nil
	}

	h := c.hostOf(id)
	if !h.nextJmp.unreachable() {
		return h.nextJmp.distance() + jmp, nil
	}

	c.discovery(ctx, h, jmp)

	return h.nextJmp.distance(), nil
}

func (c *routerImpl[Msg]) AddHosts(nodes []string) {
	topologySnapshot := c.network.Load()

	m, changed := sum(topologySnapshot.hosts, nodes)
	for changed && !c.network.CompareAndSwap(topologySnapshot, &network{
		version: topologySnapshot.version + 1,
		hosts:   m,
	}) {
		m, changed = sum(topologySnapshot.hosts, nodes)
		topologySnapshot = c.network.Load()
	}
}

func (c *routerImpl[Msg]) SetDirectly(nodes []string) {
	c.AddHosts(nodes)

	connected := make(map[string]*routeHost)
	for _, id := range nodes {
		connected[id] = c.hostOf(id)
	}

	c.direct = connected
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
func (c *routerImpl[Msg]) removeNode(node string) {
	topologySnapshot := c.network.Load()
	if _, ok := topologySnapshot.hosts[node]; !ok {
		return
	}

	m := make(map[string]byte)
	for s := range topologySnapshot.hosts {
		m[s] = 0
	}
	delete(m, node)

	for !c.network.CompareAndSwap(topologySnapshot, &network{
		version: topologySnapshot.version + 1,
		hosts:   m,
	}) {
		topologySnapshot = c.network.Load()
		if _, ok := topologySnapshot.hosts[node]; !ok {
			return
		}

		m = make(map[string]byte)
		for s := range topologySnapshot.hosts {
			m[s] = 0
		}
		delete(m, node)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.hosts, node)
	delete(c.direct, node)
}

// discovery 尝试发现目标节点
// jmp 查询请求已经经过的节点数量
func (c *routerImpl[Msg]) discovery(ctx context.Context, target *routeHost, jmp uint32) {
	// 本机节点直接返回
	if target.id == c.processor.LocalId() {
		target.nextJmp.Store(&nextJmpHost{host: target, distance: 0})
		return
	}

	// 上锁，防止同时进行冗余查询，浪费系统资源
	target.lock.Lock()
	defer target.lock.Unlock()

	// double check，确认是否已经找到目标节点
	if !target.nextJmp.unreachable() {
		return
	}

	if _, ok := ctx.Deadline(); !ok {
		//goland:noinspection GoVetLostCancel
		ctx, _ = context.WithTimeout(ctx, 45*time.Second)
	}

	// direct hosts snapshot
	direct := c.direct

	finishedCount := new(atomic.Int32)
	done := make(lang.RawChannel[bool])

	// 遍历已连接节点，发送寻人启事
	for _, next := range direct {
		njn := &nextJmpHost{
			host:     next,
			distance: 1,
		}

		// 发现下一跳正好是目标节点，立即存储
		// 但是由于分布式环境充满不确定性，还是要发送确认消息的
		if next.id == target.id {
			target.nextJmp.Store(njn)
		}

		// 启动新 goroutine 发送请求
		go c.check(ctx, next, target, jmp, func(node string, distance uint32, e exceptions.Exception) {
			if distance == UNREACHABLE || e != nil {
				target.nextJmp.CompareAndSwap(njn, unreachable)

				// 确保在所有节点都返回不可达时，能够正确返回
				if finishedCount.Add(1) == int32(len(direct)) {
					done.TrySend(true)
				}

				if e != nil {
					log.Warnf("failed to check for routeHost %s: %s", next.id, e)
					e.PrintStackTrace()
				}
			}

			target.nextJmp.update(&nextJmpHost{
				host:     next,
				distance: distance,
			})

			done.TrySend(true)
		})
	}

	if _, ok := done.ReceiveTimeout(time.Minute); !ok {
		log.Warnf("discovery to %s timeout", target.id)
	}
}

func (c *routerImpl[Msg]) check(
	ctx context.Context,
	nextJmp, target *routeHost,
	jmp uint32,
	done func(node string, distance uint32, e exceptions.Exception),
) {
	distance, e := c.processor.Find(ctx, nextJmp.id, target.id, jmp)
	if e != nil {
		done(target.id, UNREACHABLE, e)
		return
	}

	done(target.id, distance, nil)
}

// hostOf 根据 id 获取 routeHost 实例
// routeHost 作为 network 中的单例实现
func (c *routerImpl[Msg]) hostOf(id string) *routeHost {
	h := c.hosts[id]
	if h != nil {
		return h
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// double check
	h = c.hosts[id]
	if h != nil {
		return h
	}

	h = &routeHost{id: id}
	c.hosts[id] = h

	h.nextJmp.Store(unreachable)

	return h
}

func (h *nextJmpReference) unreachable() bool {
	nextJmp := h.Load()
	return nextJmp == nil || nextJmp.distance == UNREACHABLE
}

func (h *nextJmpReference) hostId() string {
	return h.Load().host.id
}

func (h *nextJmpReference) distance() uint32 {
	return h.Load().distance
}

func (h *nextJmpReference) update(nextJmp *nextJmpHost) {
	if nextJmp.distance == UNREACHABLE {
		return
	}

	snap := h.Load()
	for (snap == nil || snap.distance > nextJmp.distance) &&
		!h.CompareAndSwap(snap, nextJmp) {

		snap = h.Load()
	}
}
