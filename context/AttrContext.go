package context

import (
	com_joinu_im_protobuf "joinu-im-node/proto"
	"reflect"
	"tursom-im/attr"
)

type AttrContext struct {
	userIdAttrKey    *attr.AttachmentKey
	userTokenAttrKey *attr.AttachmentKey
}

func (a AttrContext) UserIdAttrKey() *attr.AttachmentKey {
	return a.userIdAttrKey
}

func (a AttrContext) UserTokenAttrKey() *attr.AttachmentKey {
	return a.userTokenAttrKey
}

func NewAttrContext() *AttrContext {
	userIdAttrKey := attr.NewAttachmentKey("userIdAttrKey", reflect.TypeOf(""))
	userTokenAttrKey := attr.NewAttachmentKey("userTokenAttrKey", reflect.TypeOf(&com_joinu_im_protobuf.ImToken{}))
	return &AttrContext{
		userIdAttrKey:    &userIdAttrKey,
		userTokenAttrKey: &userTokenAttrKey,
	}
}
