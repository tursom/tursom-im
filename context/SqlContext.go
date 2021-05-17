package context

import "database/sql"

type SqlContext interface {
	GetDB() *sql.DB
	GetUserTableContext() *UserTableContext
}

type User struct {
	id    string
	token []string
}

type Table interface {
	CreateTable()
}

type UserTableContext interface {
	CreateUser() *User
	FindById(uid string) *User
	GetToken(uid string) *[]string
	PushToken(uid string, token string)
}
