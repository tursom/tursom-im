package context

import (
	"database/sql"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"
)

type SqlContext interface {
	lang.Object
	GetDB() *sql.DB
	GetUserTableContext() UserTableContext
	Init(ctx *GlobalContext)
}

type Table interface {
	lang.Object
	CreateTable() exceptions.Exception
}

type UserTableContext interface {
	lang.Object
	Table
	CreateUser() (*User, exceptions.Exception)
	FindById(uid string) (*User, exceptions.Exception)
	GetToken(uid string) (*[]string, exceptions.Exception)
	PushToken(uid string, token string) exceptions.Exception
}

type User struct {
	lang.BaseObject
	id    string
	token []string
}

func (u *User) Id() string {
	return u.id
}

func (u *User) Token() []string {
	return u.token
}
