package context

import (
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom/tursom-im/conn"
	"github.com/tursom/tursom-im/proto/msys"
)

type AttrContext struct {
	lang.BaseObject
	userIdAttrKey    *conn.AttachmentKey[lang.String]
	userTokenAttrKey *conn.AttachmentKey[*msys.ImToken]
}

func (a AttrContext) UserId(c conn.Conn) conn.Attachment[lang.String] {
	return a.userIdAttrKey.Get(c)
}

func (a AttrContext) UserIdAttrKey() *conn.AttachmentKey[lang.String] {
	return a.userIdAttrKey
}

func (a AttrContext) UserToken(c conn.Conn) conn.Attachment[*msys.ImToken] {
	return a.userTokenAttrKey.Get(c)
}

func (a AttrContext) UserTokenAttrKey() *conn.AttachmentKey[*msys.ImToken] {
	return a.userTokenAttrKey
}

func NewAttrContext() *AttrContext {
	userIdAttrKey := conn.NewAttachmentKey[lang.String]("userIdAttrKey")
	userTokenAttrKey := conn.NewAttachmentKey[*msys.ImToken]("userTokenAttrKey")
	return &AttrContext{
		userIdAttrKey:    userIdAttrKey,
		userTokenAttrKey: userTokenAttrKey,
	}
}
