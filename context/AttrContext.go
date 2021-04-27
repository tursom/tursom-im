package context

import (
	"reflect"
	"tursom-im/im_conn"
	tursom_im_protobuf "tursom-im/proto"
)

type AttrContext struct {
	userIdAttrKey    *im_conn.AttachmentKey
	userTokenAttrKey *im_conn.AttachmentKey
}

func (a AttrContext) UserIdAttrKey() *im_conn.AttachmentKey {
	return a.userIdAttrKey
}

func (a AttrContext) UserTokenAttrKey() *im_conn.AttachmentKey {
	return a.userTokenAttrKey
}

func NewAttrContext() *AttrContext {
	userIdAttrKey := im_conn.NewAttachmentKey("userIdAttrKey", reflect.TypeOf(""))
	userTokenAttrKey := im_conn.NewAttachmentKey("userTokenAttrKey", reflect.TypeOf(&tursom_im_protobuf.ImToken{}))
	return &AttrContext{
		userIdAttrKey:    &userIdAttrKey,
		userTokenAttrKey: &userTokenAttrKey,
	}
}
