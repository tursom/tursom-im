package context

import "database/sql"

type SqlContext interface {
	GetDB() *sql.DB
	GetUserTableContext() UserTableContext
	Init(ctx *GlobalContext)
}

type Table interface {
	CreateTable() error
}

type UserTableContext interface {
	CreateUser() (*User, error)
	FindById(uid string) (*User, error)
	GetToken(uid string) (*[]string, error)
	PushToken(uid string, token string) error
}

type User struct {
	id    string
	token []string
}

func (u *User) Id() string {
	return u.id
}

func (u *User) Token() []string {
	return u.token
}
