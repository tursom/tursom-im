package context

import (
	b64 "encoding/base64"
	"fmt"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/exception"
	"github.com/tursom-im/tursom_im_protobuf"
)

type TokenContext struct {
	lang.BaseObject
	sqlContext SqlContext
}

func NewTokenContext() *TokenContext {
	return &TokenContext{}
}

func (c *TokenContext) Init(ctx *GlobalContext) {
	c.init(ctx.sqlContext)
}

func (c *TokenContext) init(sqlContext SqlContext) {
	c.sqlContext = sqlContext
}

func (c *TokenContext) Parse(tokenStr string) (*tursom_im_protobuf.ImToken, exceptions.Exception) {
	tokenBytes, err := b64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, exceptions.Package(err)
	}
	token := tursom_im_protobuf.ImToken{}
	err = proto.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, exceptions.Package(err)
	}

	user, err := c.sqlContext.GetUserTableContext().FindById(token.Uid)
	if err != nil {
		return nil, exceptions.Package(err)
	}
	if user == nil {
		return nil, exception.NewTokenParseException(token.Uid + " not found")
	}

	for i := range user.Token() {
		if user.Token()[i] == tokenStr {
			return &token, nil
		}
	}
	return nil, exception.NewTokenSigException(fmt.Sprintf(
		"token \"%s\" for user \"%s\" have wrong sig",
		tokenStr,
		token.Uid,
	))
}

func (c *TokenContext) FlushToken(uid string) (string, exceptions.Exception) {
	sig := rand.Uint64()
	token := &tursom_im_protobuf.ImToken{
		Uid: uid,
		Sig: sig,
	}

	bytes, err := proto.Marshal(token)
	if err != nil {
		return "", exceptions.Package(err)
	}
	newToken := b64.StdEncoding.EncodeToString(bytes)
	err = c.sqlContext.GetUserTableContext().PushToken(uid, newToken)
	if err != nil {
		return "", exceptions.Package(err)
	}
	return newToken, nil
}
