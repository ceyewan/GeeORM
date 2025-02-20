package clause

import (
	"strings"
)

// Clause 结构体用于构建 SQL 语句和对应的参数
type Clause struct {
	sql     map[Type]string        // 存储不同类型的 SQL 语句
	sqlVars map[Type][]interface{} // 存储 SQL 语句对应的参数
}

type Type int

// 定义 SQL 语句的类型
const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

// Set 方法用于设置某种类型的 SQL 语句及其对应的参数
//
// 参数:
// name: SQL 语句的类型
// vars: SQL 语句对应的参数
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// Build 方法用于根据指定的顺序构建最终的 SQL 语句和对应的参数
//
// 参数:
// orders: SQL 语句的类型顺序
//
// 返回值:
// string: 构建的 SQL 语句
// []interface{}: SQL 语句对应的参数
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
