package session

import (
	"database/sql"
	"geeorm/clause"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/schema"
	"strings"
)

// Session 是会话管理的主要结构，包含会话的所有操作
type Session struct {
	db       *sql.DB         // 数据库连接
	sql      strings.Builder // sql 用于拼接 SQL 语句
	sqlVars  []interface{}   // sqlVars 用于存储 SQL 语句中的参数
	dialect  dialect.Dialect // dialect 记录了该 Session 所使用的数据库方言
	refTable *schema.Schema  // refTable 记录 Model 对应的表结构
	clause   clause.Clause   // clause 是记录 SQL 语句中的各种子句
	tx       *sql.Tx         // tx 提供事务支持，如果 tx 不为 nil，则执行所有操作都在事务中
}

// New 返回一个新的会话
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{db: db,
		dialect: dialect,
	}
}

// Clear 重置 Session 中的 SQL 语句和参数列表
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

// 抽象出一个接口 CommonDB，包含 Query、QueryRow、Exec 三个方法
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// 该接口的实现有 *sql.DB 和 *sql.Tx
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB 返回 *sql.DB 对象
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// Raw 方法用于 SQL 语句
//
// 参数：
// sql: 用来拼接的 SQL 语句
// values: SQL 语句中占位符的值
//
// 返回值：
// *Session: 返回 Session 实例
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec 执行 s.sql 这条 SQL 语句，参数为 s.sqlVars
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow 执行 s.sql 这条 SQL 语句，参数为 s.sqlVars
// 并且返回一行记录，该记录是 *sql.Row 类型
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.db.QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 执行 s.sql 这条 SQL 语句，参数为 s.sqlVars
// 并且返回多行记录，该记录是 *sql.Rows 类型
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// Exec 方法用于执行不需要返回行的 SQL 语句，例如 INSERT、DELETE、UPDATE 等
// 返回值 result 是 sql.Result 类型，用于返回执行结果（受影响的行数等）
// QueryRow 方法用于执行需要返回单行结果的 SQL 查询，例如 SELECT 查询
// 返回一个 *sql.Row 对象，表示查询结果的单行记录
