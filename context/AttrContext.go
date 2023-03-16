package context

import (
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/conn"
	"github.com/tursom-im/proto/encode"
)

type AttrContext struct {
	lang.BaseObject
	userIdAttrKey    *conn.AttachmentKey[lang.String]
	userTokenAttrKey *conn.AttachmentKey[*encode.ImToken]
}

func (a AttrContext) UserIdAttrKey() *conn.AttachmentKey[lang.String] {
	return a.userIdAttrKey
}

func (a AttrContext) UserTokenAttrKey() *conn.AttachmentKey[*encode.ImToken] {
	return a.userTokenAttrKey
}

func NewAttrContext() *AttrContext {
	userIdAttrKey := conn.NewAttachmentKey[lang.String]("userIdAttrKey")
	userTokenAttrKey := conn.NewAttachmentKey[*encode.ImToken]("userTokenAttrKey")
	return &AttrContext{
		userIdAttrKey:    userIdAttrKey,
		userTokenAttrKey: userTokenAttrKey,
	}
}
