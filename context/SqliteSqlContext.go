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
	db, err := sql.Open("sqlite3", "im.db")
	if err != nil {
		exceptions.Print(err)
		return nil
	}
	db.SetMaxOpenConns(1)
	s := &SqliteSqlContext{
		db:               db,
		userTableContext: NewSqliteUserTableContext(db),
	}
	err = s.userTableContext.CreateTable()
	if err != nil {
		exceptions.Print(err)
		return nil
	}
	return s
}

func (s *SqliteSqlContext) Init(ctx *GlobalContext) {
	s.userTableContext.Init(ctx)
}

func (s *SqliteSqlContext) GetDB() *sql.DB {
	return s.db
}

func (s *SqliteSqlContext) GetUserTableContext() UserTableContext {
	return s.userTableContext
}
