package conn

import (
	"io"
	"unsafe"

	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
	"github.com/tursom/GoCollections/lang/atomic"
	unsafe2 "github.com/tursom/GoCollections/unsafe"

	"github.com/tursom-im/proto/pkg"
)

var attachmentKeyId = atomic.Int32(0)

type (
	Conn interface {
		lang.Object
		io.Writer
		SendMsg(msg *pkg.ImMsg)
		WriteData(data []byte)
		Attr(*AttachmentKey[any]) Attachment[any]
		AddEventListener(func(event Event)) EventListener
	}

	AttachmentKey[T any] struct {
		lang.BaseObject
		name string
		id   int32
	}

	Attachment[T any] interface {
		lang.Object
		Get() T
		TryGet() (T, bool)
		Set(value T)
	}

	// attachmentPackager package Attachment returns from Conn.Attr
	attachmentPackager[T any] struct {
		lang.BaseObject
		attr Attachment[any]
	}
)

func NewAttachmentKey[T any](name string) *AttachmentKey[T] {
	return &AttachmentKey[T]{
		name: name,
		id:   attachmentKeyId.Add(1),
	}
}

func (a *AttachmentKey[T]) Name() string {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentKey is null", nil))
	}
	return a.name
}

func (a *AttachmentKey[T]) Id() int32 {
	if a == nil {
		panic(exceptions.NewNPE("AttachmentKey is null", nil))
	}
	return a.id
}

func (a *AttachmentKey[T]) Get(c Conn) Attachment[T] {
	attr := c.Attr(unsafe2.ForceCast[AttachmentKey[any]](unsafe.Pointer(a)))
	return &attachmentPackager[T]{attr: attr}
}

func (a *attachmentPackager[T]) Get() T {
	return a.attr.Get().(T)
}

func (a *attachmentPackager[T]) TryGet() (T, bool) {
	get, ok := a.attr.TryGet()
	if !ok {
		return lang.Nil[T](), false
	}

	return get.(T), true
}

func (a *attachmentPackager[T]) Set(value T) {
	a.attr.Set(value)
}
