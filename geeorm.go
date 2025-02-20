package geeorm

import (
	"database/sql"
	"fmt"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"
	"strings"
)

// Engine 是 GeeORM 的核心结构体，负责数据库连接管理和会话创建
type Engine struct {
	db      *sql.DB         // 数据库连接
	dialect dialect.Dialect // 数据库方言
}

// NewEngine 创建一个新的 Engine 实例
//
// 参数:
// driverName: 数据库驱动名称
// dataSourceName: 数据库连接字符串
//
// 返回值:
// *Engine: 返回创建的 Engine 实例
// error: 如果创建过程中发生错误，返回错误信息
func NewEngine(driverName, dataSourceName string) (e *Engine, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Error(err)
		return
	}
	// 发送 ping 确保数据库连接是活跃的
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// 确保指定的数据库方言已注册
	dial, ok := dialect.GetDialect(driverName)
	if !ok {
		log.Error("dialect %s Not Found", driverName)
		return
	}
	// 创建 Engine 实例并返回
	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return
}

// Close 关闭数据库连接
func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

// NewSession 创建一个新的 Session 实例
func (e *Engine) NewSession() *session.Session {
	return session.New(e.db, e.dialect)
}

// TxFunc 用于执行事务的函数，接收一个 Session 实例作为参数，返回一个结果和一个错误
type TxFunc func(*session.Session) (interface{}, error)

// Transaction 用于执行一个事务
//
// 参数:
// f: 事务函数
//
// 返回值:
// interface{}: 事务函数的返回值
// error: 如果事务执行过程中发生错误，返回错误信息
//
// 开启事务后，通过 defer 确保事务的正确结束
func (e *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := e.NewSession()
	if err = s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback() // 回滚事务
			panic(p)         // 在回滚之后继续抛出 panic
		} else if err != nil {
			_ = s.Rollback() // 回滚事务
		} else {
			err = s.Commit() // 提交事务
		}
	}()
	// 执行事务函数
	return f(s)
}

// difference 比较两个字符串切片的差异，返回 a - b 的结果
func difference(a, b []string) (diff []string) {
	m := make(map[string]bool)
	for _, s := range b {
		m[s] = true
	}
	for _, s := range a {
		if _, ok := m[s]; !ok {
			diff = append(diff, s)
		}
	}
	return
}

// Migrate 根据 value 的类型创建表结构
func (engine *Engine) Migrate(value interface{}) error {
	// 事务操作
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		// 如果表不存在，则创建表
		if !s.Model(value).HasTable() {
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		// 获取表结构
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v, deleted cols %v", addCols, delCols)
		// 如果表存在，但字段不一致，则修改表结构
		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}
		// 删除表中不存在的字段
		if len(delCols) == 0 {
			return
		}
		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
