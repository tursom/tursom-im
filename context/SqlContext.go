package context

import (
	"database/sql"
	"github.com/tursom/GoCollections/exceptions"
)

type SqlContext interface {
	GetDB() *sql.DB
	GetUserTableContext() UserTableContext
	Init(ctx *GlobalContext)
}

type Table interface {
	CreateTable() exceptions.Exception
}

type UserTableContext interface {
	CreateUser() (*User, exceptions.Exception)
	FindById(uid string) (*User, exceptions.Exception)
	GetToken(uid string) (*[]string, exceptions.Exception)
	PushToken(uid string, token string) exceptions.Exception
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
