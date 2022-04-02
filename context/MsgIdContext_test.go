package context

import (
	"fmt"
	"testing"
)

func TestMsgIdContext_NewMsgIdStr(t *testing.T) {
	ctx := NewMsgIdContext()
	fmt.Println(ctx.NewMsgIdStr())
}
