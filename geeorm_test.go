package geeorm

import (
	"errors"
	"geeorm/log"
	"geeorm/session"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine("sqlite3", "test.db")
	if err != nil {
		t.Fatal("failed to connect database:", err)
	}
	defer engine.Close()
}

func TestEngine_Close(t *testing.T) {
	engine, err := NewEngine("sqlite3", "test.db")
	if err != nil {
		t.Fatal("failed to connect database:", err)
	}
	engine.Close()
}

func TestEngine_NewSession(t *testing.T) {
	engine, err := NewEngine("sqlite3", "test.db")
	if err != nil {
		t.Fatal("failed to connect database:", err)
	}
	defer engine.Close()

	s := engine.NewSession()
	if s == nil {
		t.Fatal("failed to create new session")
	}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel) // 设置日志级别为 Error，减少测试输出
	m.Run()
}

func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatal("failed to connect", err)
	}
	return engine
}

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}

func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Tom", 18})
		return nil, errors.New("Error")
	})
	if err == nil || s.HasTable() {
		t.Fatal("failed to rollback")
	}
}

func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Tom", 18})
		return
	})
	u := &User{}
	_ = s.First(u)
	if err != nil || u.Name != "Tom" {
		t.Fatal("failed to commit")
	}
}
