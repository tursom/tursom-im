package context

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type SqliteSqlContext struct {
	lang.BaseObject
	db               *sql.DB
	userTableContext *SqliteUserTableContext
}

func NewSqliteSqlContext() *SqliteSqlContext {
	db := exceptions.Exec2r1(sql.Open, "sqlite3", "im.db")
	db.SetMaxOpenConns(1)
	s := &SqliteSqlContext{
		db:               db,
		userTableContext: NewSqliteUserTableContext(db),
	}
	exceptions.Exec0r0(s.userTableContext.CreateTable)
	return s
}

func (s *SqliteSqlContext) Init(ctx *GlobalContext) {
	s.init(ctx.msgIdContext)
}

func (s *SqliteSqlContext) init(msgIdContext *MsgIdContext) {
	s.userTableContext.init(msgIdContext)
}

func (s *SqliteSqlContext) GetDB() *sql.DB {
	return s.db
}

func (s *SqliteSqlContext) GetUserTableContext() UserTableContext {
	return s.userTableContext
}
