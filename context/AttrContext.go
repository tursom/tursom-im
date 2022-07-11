package context

import (
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/im_conn"
	"github.com/tursom-im/tursom_im_protobuf"
)

type AttrContext struct {
	lang.BaseObject
	userIdAttrKey    *im_conn.AttachmentKey[lang.String]
	userTokenAttrKey *im_conn.AttachmentKey[*tursom_im_protobuf.ImToken]
}

func (a AttrContext) UserIdAttrKey() *im_conn.AttachmentKey[lang.String] {
	return a.userIdAttrKey
}

func (a AttrContext) UserTokenAttrKey() *im_conn.AttachmentKey[*tursom_im_protobuf.ImToken] {
	return a.userTokenAttrKey
}

func NewAttrContext() *AttrContext {
	userIdAttrKey := im_conn.NewAttachmentKey[lang.String]("userIdAttrKey")
	userTokenAttrKey := im_conn.NewAttachmentKey[*tursom_im_protobuf.ImToken]("userTokenAttrKey")
	return &AttrContext{
		userIdAttrKey:    &userIdAttrKey,
		userTokenAttrKey: &userTokenAttrKey,
	}
}
