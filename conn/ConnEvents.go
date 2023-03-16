package conn

import (
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type EventId int32

const (
	ConnClosedId EventId = 1
)

type (
	Event interface {
		lang.Object
		EventId() EventId
		Conn() Conn
	}

	AbstractEvent struct {
		lang.BaseObject
		eventId EventId
		conn    Conn
	}

	EventClosed struct {
		AbstractEvent
	}

	EventListener interface {
		Remove() exceptions.Exception
	}
)

func (a *AbstractEvent) EventId() EventId {
	return a.eventId
}

func (a *AbstractEvent) Conn() Conn {
	return a.conn
}

func NewAbstractEvent(eventId EventId, conn Conn) *AbstractEvent {
	return &AbstractEvent{eventId: eventId, conn: conn}
}

func NewEventClosed(conn Conn) *EventClosed {
	return &EventClosed{AbstractEvent: *NewAbstractEvent(ConnClosedId, conn)}
}

func (i EventId) IsConnClosed() bool {
	return i == ConnClosedId
}
