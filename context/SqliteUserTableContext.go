package context

import (
	"database/sql"
	"encoding/json"
	"github.com/tursom/GoCollections/exceptions"
	"tursom-im/exception"
)

type SqliteUserTableContext struct {
	db           *sql.DB
	msgIdContext *MsgIdContext
}

type UserNotFoundError struct {
	uid string
}

func (u *UserNotFoundError) Error() string {
	return "user \"" + u.uid + "\" not found"
}

func (s *SqliteUserTableContext) Init(ctx *GlobalContext) {
	s.msgIdContext = ctx.msgIdContext
}

func (s *SqliteUserTableContext) CreateTable() error {
	_, err := s.db.Exec("create table if not exists user(" +
		"	id char(32) primary key not null," +
		"	token text" +
		")")
	return exceptions.Package(err)
}

func (s *SqliteUserTableContext) CreateUser() (*User, error) {
	newUserId := s.msgIdContext.NewMsgIdStr()
	_, err := s.db.Exec("insert into user (id,token) values (?,?)", newUserId, "[]")
	if err != nil {
		return nil, exceptions.Package(err)
	}
	return s.FindById(newUserId)
}

func (s *SqliteUserTableContext) CreateUserWithToken(uid string, token string) (*User, error) {
	_, err := s.db.Exec("insert into user (id,token) values (?,?)", uid, "[\""+token+"\"]")
	if err != nil {
		return nil, exceptions.Package(err)
	}
	return s.FindById(uid)
}

func (s *SqliteUserTableContext) FindById(uid string) (*User, error) {
	rows, err := s.db.Query("select id,token from user where id=?", uid)
	if err != nil {
		return nil, exceptions.Package(err)
	}
	defer func() {
		exceptions.Print(rows.Close())
	}()
	if !rows.Next() {
		return nil, exception.NewUserNotFoundException("user " + uid + " not found")
	}

	user := &User{}
	var token string
	err = rows.Scan(&user.id, &token)
	if err != nil {
		return nil, exceptions.Package(err)
	}
	err = json.Unmarshal([]byte(token), &user.token)
	if err != nil {
		return nil, exceptions.Package(err)
	}
	return user, nil
}

func (s *SqliteUserTableContext) GetToken(uid string) (*[]string, error) {
	user, err := s.FindById(uid)
	if user == nil || err != nil {
		return nil, exceptions.Package(err)
	}
	return &user.token, nil
}

func (s *SqliteUserTableContext) PushToken(uid string, token string) error {
	tokenExist, err := s.GetToken(uid)
	if err != nil {
		return err
	}
	if tokenExist == nil {
		return &UserNotFoundError{uid}
	}

	newToken := []string{token}
	for i := 0; i < 9 && i < len(*tokenExist); i++ {
		newToken = append(newToken, (*tokenExist)[i])
	}

	tokenBytes, err := json.Marshal(newToken)
	if err != nil {
		return exceptions.Package(err)
	}

	_, err = s.db.Exec("update user set token=? where id = ?", string(tokenBytes), uid)
	return exceptions.Package(err)
}

func NewSqliteUserTableContext(db *sql.DB) *SqliteUserTableContext {
	return &SqliteUserTableContext{db: db}
}
