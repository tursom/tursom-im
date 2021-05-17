package context

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteSqlContext struct {
	db               *sql.DB
	userTableContext SqliteUserTableContext
}

func NewSqliteSqlContext() *SqliteSqlContext {
	db, err := sql.Open("sqlite3", "im.db")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	s := &SqliteSqlContext{
		db:               db,
		userTableContext: SqliteUserTableContext{db: db},
	}
	s.userTableContext.CreateTable()
	return s
}

func (s *SqliteSqlContext) Init(ctx *GlobalContext) {
	s.userTableContext.Init(ctx)
}

func (s *SqliteSqlContext) GetDB() *sql.DB {
	return s.db
}

func (s *SqliteSqlContext) GetUserTableContext() *UserTableContext {
	panic("implement me")
}
