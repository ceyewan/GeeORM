package dialect

import "reflect"

// dialectsMap 存储所有注册的数据库方言
var dialectsMap = map[string]Dialect{}

// Dialect 是一个接口，定义了数据库方言需要实现的方法
type Dialect interface {
	// DataTypeOf 返回 Go 语言类型在数据库中的数据类型
	//
	// 参数:
	// typ: Go 语言的反射类型
	//
	// 返回值:
	// string: 数据库中的数据类型
	DataTypeOf(typ reflect.Value) string

	// TableExistSQL 返回检查表是否存在的 SQL 语句
	//
	// 参数:
	// tableName: 表名
	//
	// 返回值:
	// string: SQL 查询语句
	// []interface{}: 查询参数
	TableExistSQL(tableName string) (string, []interface{})
}

// RegisterDialect 注册一个数据库方言
//
// 参数:
// name: 数据库方言名称
// dialect: 实现了 Dialect 接口的数据库方言实例
func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

// GetDialect 获取指定名称的数据库方言
//
// 参数:
// name: 数据库方言名称
//
// 返回值:
// Dialect: 返回对应的数据库方言
// bool: 如果方言存在，返回 true；否则返回 false
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
