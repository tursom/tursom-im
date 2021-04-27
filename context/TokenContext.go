package context

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"tursom-im/proto"
)

type TokenContext struct {
}

func (receiver TokenContext) Parse(tokenStr string) (*tursom_im_protobuf.ImToken, error) {
	tokenBytes, err := b64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	token := tursom_im_protobuf.ImToken{}
	err = proto.Unmarshal(tokenBytes, &token)
	return &token, err
}
