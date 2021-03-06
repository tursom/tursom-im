package context

import (
	"database/sql"
	"encoding/json"

	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom-im/exception"
)

const (
	tableUserCreate = `create table if not exists user(
	id char(32) primary key not null,
	token text
)`
	rowUserCreate = `insert into user (id,token) values (?,?)`
	rowUserQuery  = `select id,token from user where id=?`
)

type (
	SqliteUserTableContext struct {
		lang.BaseObject
		db           *sql.DB
		msgIdContext *MsgIdContext
	}

	UserNotFoundError struct {
		exceptions.RuntimeException
		uid string
	}
)

func (u *UserNotFoundError) Error() string {
	return "user \"" + u.uid + "\" not found"
}

func (s *SqliteUserTableContext) Init(ctx *GlobalContext) {
	s.msgIdContext = ctx.msgIdContext
}

func (s *SqliteUserTableContext) init(msgIdContext *MsgIdContext) {
	s.msgIdContext = msgIdContext
}

func (s *SqliteUserTableContext) CreateTable() exceptions.Exception {
	_, err := s.db.Exec(tableUserCreate)
	return exceptions.Package(err)
}

func (s *SqliteUserTableContext) CreateUser() (*User, exceptions.Exception) {
	newUserId := s.msgIdContext.NewMsgIdStr()
	_, err := s.db.Exec(rowUserCreate, newUserId, "[]")
	if err != nil {
		return nil, exceptions.Package(err)
	}
	return s.FindById(newUserId)
}

func (s *SqliteUserTableContext) CreateUserWithToken(uid string, token string) (*User, exceptions.Exception) {
	_, err := s.db.Exec(rowUserCreate, uid, "[\""+token+"\"]")
	if err != nil {
		return nil, exceptions.Package(err)
	}
	return s.FindById(uid)
}

func (s *SqliteUserTableContext) FindById(uid string) (*User, exceptions.Exception) {
	rows, err := s.db.Query(rowUserQuery, uid)
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

func (s *SqliteUserTableContext) GetToken(uid string) (*[]string, exceptions.Exception) {
	user, err := s.FindById(uid)
	if user == nil || err != nil {
		return nil, exceptions.Package(err)
	}
	return &user.token, nil
}

func (s *SqliteUserTableContext) PushToken(uid string, token string) exceptions.Exception {
	tokenExist, err := s.GetToken(uid)
	if err != nil {
		return err
	}
	if tokenExist == nil {
		return exception.NewUserNotFoundException(uid)
	}

	newToken := []string{token}
	for i := 0; i < 9 && i < len(*tokenExist); i++ {
		newToken = append(newToken, (*tokenExist)[i])
	}

	tokenBytes, err2 := json.Marshal(newToken)
	if err2 != nil {
		return exceptions.Package(err2)
	}

	_, err2 = s.db.Exec("update user set token=? where id = ?", string(tokenBytes), uid)
	return exceptions.Package(err2)
}

func NewSqliteUserTableContext(db *sql.DB) *SqliteUserTableContext {
	return &SqliteUserTableContext{db: db}
}
