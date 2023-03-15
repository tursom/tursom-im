package distributed

import (
	"sync"

	"github.com/tursom/GoCollections/exceptions"
)

type (
	Local interface {
		IsLocal(id string) bool
		UserNodeId(id string) string
		Send(id string, msg []byte) exceptions.Exception
	}

	Topology struct {
		localHandler Local
		version      uint32
		nodeSet      map[string]*Node
		users        map[string]*User
		lock         sync.RWMutex
	}

	Node struct {
		id      string
		route   map[*Node]byte
		nextJmp *Node
		state   NoswState
	}

	User struct {
		distance uint32
		id       string
		node     *Node
		lock     sync.RWMutex
	}

	NoswState int8
)

// NoswState
const (
	NodeStateUnknown NoswState = iota
	NodeStateDiscovering
	NodeFounded
	NodeOffline
	NodeStateLocal
)

func (t *Topology) Send(id string, msg []byte) exceptions.Exception {
	// TODO valid id format
	user := t.getUser(id)

	switch user.state {
	case NodeStateUnknown:
		if t.localHandler.IsLocal(id) {
			user.state = NodeStateLocal
			return t.localHandler.Send(id, msg)
		}

		if e := t.discovery(t.localHandler.UserNodeId(id)); e != nil {
			return e
		}

	case NodeStateDiscovering:
	case NodeFounded:
	case NodeStateLocal:
		return t.localHandler.Send(id, msg)
	}
}

func (t *Topology) discovery(node string) exceptions.Exception {
	func() {
		t.lock.Lock()
		defer t.lock.Unlock()

	}()
}

func (t *Topology) getUser(id string) *User {
	user := t.users[id]
	if user != nil {
		return user
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	// double check
	user = t.users[id]
	if user != nil {
		return user
	}

	user = new(User)
	t.users[id] = user

	return user
}
