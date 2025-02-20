package dialect

import (
	"fmt"
	"reflect"
	"time"
)

// sqlite3 是一个实现了 Dialect 接口的结构体，用于处理 SQLite3 数据库的方言
type sqlite3 struct{}

// 类型断言，确保 sqlite3 实现了 Dialect 接口
var _ Dialect = (*sqlite3)(nil)

// init 函数在包被导入时自动执行，用于注册 sqlite3 数据库方言
func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

// DataTypeOf 返回 Go 语言类型在 SQLite3 数据库中的数据类型
//
// 参数:
// typ: Go 语言的反射类型
//
// 返回值:
// string: 数据库中的数据类型
func (s *sqlite3) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		// typ.Interface() 返回接口的动态值, 类型为 interface{}，也即空接口，any 类型
		// 如果 typ.Interface() 是 time.Time 类型（类型断言），则返回 "datetime"
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

// TableExistSQL 生成检查 SQLite 数据库中某个表是否存在的 SQL 语句
//
// 参数:
// tableName: 表名
//
// 返回值:
// string: SQL 查询语句
// []interface{}: 查询参数
func (s *sqlite3) TableExistSQL(tableName string) (string, []interface{}) {
	// 将表名作为参数传入 SQL 查询语句
	args := []interface{}{tableName}
	// sqlite_master 表包含了数据库中的所有表、索引、视图和触发器的信息
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}
