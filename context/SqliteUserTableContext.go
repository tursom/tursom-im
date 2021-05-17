package context

import (
	"database/sql"
	"fmt"
)

type SqliteUserTableContext struct {
	db           *sql.DB
	msgIdContext *MsgIdContext
}

func (s *SqliteUserTableContext) Init(ctx *GlobalContext) {
	s.msgIdContext = ctx.msgIdContext
}

func (s *SqliteUserTableContext) CreateTable() {
	s.db.Exec("create table if not exists user(" +
		"	id char(32) primary key not null," +
		"	token text" +
		")")
}

func (s *SqliteUserTableContext) CreateUser() *User {
	newUserId := s.msgIdContext.NewMsgIdStr()
	s.db.Exec("insert into user (id,token) values (?,?)", newUserId, "[]")
	return s.FindById(newUserId)
}

func (s *SqliteUserTableContext) FindById(uid string) *User {
	rows, err := s.db.Query("select id,token from user where id=?", uid)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()
	if !rows.Next() {
		return nil
	}

	panic("implement me")
}

func (s *SqliteUserTableContext) GetToken(uid string) *[]string {
	user := s.FindById(uid)
	if user == nil {
		return nil
	}
	return &user.token
}

func (s *SqliteUserTableContext) PushToken(uid string, token string) {
	panic("implement me")
}

func NewSqliteUserTableContext(db *sql.DB) *SqliteUserTableContext {
	return &SqliteUserTableContext{db: db}
}
