package context

import (
	"database/sql"

	"gitea.tursom.cn/tursom/kvs/kv"
	"gitea.tursom.cn/tursom/kvs/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tursom/GoCollections/exceptions"
	"github.com/tursom/GoCollections/lang"

	"github.com/tursom/tursom-im/proto/ie"
)

var (
	tableVersionKey = ie.SysKey("sys:tableVersion")

	tableSqls = []string{
		`create table if not exists user(
			id char(32) primary key not null,
			token text
		)`,
	}
)

type SqliteSqlContext struct {
	lang.BaseObject
	db               *sql.DB
	kvContext        kv.Store[*ie.KVStoreKey, []byte]
	stringKvContext  kv.Store[*ie.KVStoreKey, string]
	uint32KvContext  kv.Store[*ie.KVStoreKey, uint32]
	userTableContext *SqliteUserTableContext
}

func NewSqliteSqlContext() *SqliteSqlContext {
	db := exceptions.Exec2r1(sql.Open, "sqlite3", "im.db")
	db.SetMaxOpenConns(1)

	s := &SqliteSqlContext{
		db: db,
		kvContext: kv.KCodecStore(
			exceptions.Exec2r1(sqlite.New, db, "kv"),
			kv.ProtoCodec(func() *ie.KVStoreKey { return &ie.KVStoreKey{} }),
		),
		userTableContext: NewSqliteUserTableContext(db),
	}
	s.stringKvContext = kv.VCodecStore(s.KVS(), kv.StringToByteCodec)
	s.uint32KvContext = kv.VCodecStore(s.KVS(), kv.Uint32ToByteCodec)

	return s
}

func (s *SqliteSqlContext) createTables() {
	version := exceptions.Exec1r1(s.uint32KvContext.Get, tableVersionKey)
	remainSqls := tableSqls[version:]
	for i, tableSql := range remainSqls {
		if _, err := s.db.Exec(tableSql); err != nil {
			panic(exceptions.Package(err))
		}

		exceptions.Exec2r0(s.uint32KvContext.Put, tableVersionKey, version+1+uint32(i))
	}
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

func (s *SqliteSqlContext) KVS() kv.Store[*ie.KVStoreKey, []byte] {
	return s.kvContext
}

func (s *SqliteSqlContext) StringKVS() kv.Store[*ie.KVStoreKey, string] {
	return s.stringKvContext
}

func (s *SqliteSqlContext) Uint32KVS() kv.Store[*ie.KVStoreKey, uint32] {
	return s.uint32KvContext
}
