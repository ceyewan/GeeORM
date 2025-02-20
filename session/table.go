// 用于放置操作数据库表相关的代码
package session

import (
	"fmt"
	"geeorm/log"
	"geeorm/schema"
	"reflect"
	"strings"
)

// Model 设置当前会话操作的模型
//
// 参数:
// value: 要操作的模型对象
//
// 返回值:
// *Session: 返回当前会话实例
func (s *Session) Model(value interface{}) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// RefTable 返回当前会话操作的表的 Schema
//
// 返回值:
// *schema.Schema: 表的 Schema 对象
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

// CreateTable 创建数据库表
//
// 返回值:
// error: 如果创建过程中发生错误，返回错误信息
func (s *Session) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

// DropTable 删除数据库表
//
// 返回值:
// error: 如果删除过程中发生错误，返回错误信息
func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

// HasTable 检查数据库表是否存在
//
// 返回值:
// bool: 如果表存在，返回 true；否则返回 false
func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	// row.Scan 将数据库中的值扫描到 tmp 中
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().Name
}
