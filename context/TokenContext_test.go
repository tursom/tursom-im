package context

import (
	"fmt"
	"testing"
)

func TestTokenContext_FlushToken(t *testing.T) {
	ctx := NewTokenContext()
	{
		sqlContext := NewSqliteSqlContext()
		sqlContext.init(NewMsgIdContext())
		ctx.init(sqlContext)
	}
	fmt.Println(ctx.FlushToken("23imignRsEU"))
}
