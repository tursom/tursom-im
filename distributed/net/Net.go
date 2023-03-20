package net

import (
	"context"
	"sync"
	"time"

	"github.com/tursom/GoCollections/exceptions"

	"github.com/tursom-im/distributed/router"
)

type (
	Processor interface {
		BroadcastNodeUnreachable(node string)
	}

	Net[Msg any] interface {
		router.Router[Msg]
		HostOnline(id string)
	}

	netImpl[Msg any] struct {
		router    router.Router[Msg]
		lock      sync.RWMutex
		blacklist map[string]time.Time
	}
)

func NewNet[Msg any](config *router.Config[Msg]) Net[Msg] {
	return NewNetWithRouter(router.NewRouter(config))
}

func NewNetWithRouter[Msg any](router router.Router[Msg]) Net[Msg] {
	return &netImpl[Msg]{
		router:    router,
		blacklist: make(map[string]time.Time),
	}
}

func (c *netImpl[Msg]) AddHosts(hosts []string) {
	filtered := c.filterBlacklist(hosts)

	c.router.AddHosts(filtered)
}

func (c *netImpl[Msg]) filterBlacklist(hosts []string) []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	filtered := make([]string, 0, len(hosts))

	for _, host := range hosts {
		if t, ok := c.blacklist[host]; ok || time.Now().Before(t) {
			continue
		}

		filtered = append(filtered, host)
	}
	return filtered
}

func (c *netImpl[Msg]) hostUnreachable(host string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.blacklist[host] = time.Now().Add(5 * time.Minute)
}

func (c *netImpl[Msg]) HostOnline(id string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.blacklist, id)
	c.router.AddHosts([]string{id})
}

func (c *netImpl[Msg]) Send(ctx context.Context, id string, msg Msg, jmp uint32) (bool, exceptions.Exception) {
	send, e := c.router.Send(ctx, id, msg, jmp)
	if !send && e == nil {
		c.hostUnreachable(id)
	}

	return send, e
}

func (c *netImpl[Msg]) Find(ctx context.Context, id string, jmp uint32) (uint32, exceptions.Exception) {
	find, e := c.router.Find(ctx, id, jmp)
	if find == router.UNREACHABLE && e == nil {
		c.hostUnreachable(id)
	}

	return find, e
}

func (c *netImpl[Msg]) SetDirectly(hosts []string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, host := range hosts {
		delete(c.blacklist, host)
	}

	c.router.SetDirectly(hosts)
}
