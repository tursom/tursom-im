package context

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	com_joinu_im_protobuf "joinu-im-node/proto"
)

type TokenContext struct {
}

func (receiver TokenContext) Parse(tokenStr string) (*com_joinu_im_protobuf.ImToken, error) {
	tokenBytes, err := b64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	token := com_joinu_im_protobuf.ImToken{}
	err = proto.Unmarshal(tokenBytes, &token)
	if err != nil {
		tokenOld := com_joinu_im_protobuf.ImTokenOld{}
		err = proto.Unmarshal(tokenBytes, &tokenOld)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		token.AppId = tokenOld.AppId
		token.AppUid = tokenOld.AppUid
		token.Uid = tokenOld.Uid
		token.Sig = tokenOld.Sig
	}
	//fmt.Println("token:", token)
	return &token, nil
}
