package context

import (
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tursom/GoCollections/exceptions"
	"math"
	"math/big"
	"tursom-im/exception"
	"tursom-im/tursom_im_protobuf"
)

type TokenContext struct {
	sqlContext SqlContext
}

type TokenSigError struct {
	token string
	uid   string
}

func (t *TokenSigError) Error() string {
	return "token \"" + t.token + "\" for user \"" + t.uid + "\" have wrong sig"
}

type TokenParseError struct {
	msg string
}

func (t *TokenParseError) Error() string {
	return t.msg
}

func NewTokenContext() *TokenContext {
	return &TokenContext{}
}

func (c *TokenContext) Init(ctx *GlobalContext) {
	c.sqlContext = ctx.sqlContext
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
	return nil, exception.NewTokenSigException(fmt.Sprintf("token \"%s\" for user \"%s\" have wrong sig", tokenStr, token.Uid))
}

func (c *TokenContext) FlushToken(uid string) (string, exceptions.Exception) {
	sig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", exceptions.Package(err)
	}
	token := &tursom_im_protobuf.ImToken{
		Uid: uid,
		Sig: sig.Uint64(),
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
