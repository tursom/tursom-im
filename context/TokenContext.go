package context

import (
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"math/big"
	"tursom-im/tursom_im_protobuf"
)

type TokenContext struct {
	sigMap     map[string]uint64
	sqlContext SqlContext
}

func NewTokenContext() *TokenContext {
	return &TokenContext{
		map[string]uint64{},
		nil,
	}
}

func (c *TokenContext) Init(ctx *GlobalContext) {
	c.sqlContext = ctx.sqlContext
}

func (c *TokenContext) Parse(tokenStr string) (*tursom_im_protobuf.ImToken, error) {
	tokenBytes, err := b64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	token := tursom_im_protobuf.ImToken{}
	err = proto.Unmarshal(tokenBytes, &token)
	return &token, err
}

func (c *TokenContext) FlushToken(uid string) (string, error) {
	sig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}
	c.sigMap[uid] = sig.Uint64()
	token := &tursom_im_protobuf.ImToken{
		Uid: uid,
		Sig: sig.Uint64(),
	}

	bytes, err := proto.Marshal(token)
	if err != nil {
		return "", err
	}
	newToken := b64.StdEncoding.EncodeToString(bytes)
	err = c.sqlContext.GetUserTableContext().PushToken(uid, newToken)
	if err != nil {
		return "", err
	}
	return newToken, nil
}
