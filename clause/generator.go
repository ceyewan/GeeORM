package clause

import (
	"fmt"
	"strings"
)

// generator 是一个函数类型，用于生成 SQL 语句和对应的参数
//
// 参数:
// values: 可变参数，用于生成 SQL 语句的值
//
// 返回值:
// string: 生成的 SQL 语句
// []interface{}: SQL 语句对应的参数
type generator func(values ...interface{}) (string, []interface{})

// generators 存储了不同类型的 SQL 语句生成器
var generators map[Type]generator

// init 用于初始化 generators
func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

// genBindVars 生成指定数量的占位符
//
// 参数:
// num: 占位符的数量
//
// 返回值:
// string: 生成的占位符字符串
//
// genBindVars(3) => "?, ?, ?"
func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

// _insert 生成 INSERT 语句
//
// 参数:
// values: 可变参数，第一个参数是表名，第二个参数是字段名列表
//
// 返回值:
// string: 生成的 INSERT 语句
// []interface{}: 空的参数列表
//
// _insert("users", []string{"Name", "Age"}) => "INSERT INTO users (Name, Age)"
func _insert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName ($fields)
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ", ")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

// _values 生成 VALUES 语句
//
// 参数:
// values: 可变参数，每个参数是一个值列表
//
// 返回值:
// string: 生成的 VALUES 语句
// []interface{}: 所有的值
//
// _values([]interface{}{"Tom", 18}, []interface{}{"Sam", 25}) => "VALUES (?, ?), (?, ?)", []interface{}{"Tom", 18, "Sam", 25}
func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ($v3)
	var bindStr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars
}

// _select 生成 SELECT 语句
//
// 参数:
// values: 可变参数，第一个参数是表名，第二个参数是字段名列表
//
// 返回值:
// string: 生成的 SELECT 语句
// []interface{}: 空的参数列表
//
// _select("users", []string{"Name", "Age"}) => "SELECT Name, Age FROM users"
func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ", ")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

// _limit 生成 LIMIT 语句
//
// 参数:
// values: 可变参数，第一个参数是限制的数量
//
// 返回值:
// string: 生成的 LIMIT 语句
// []interface{}: 限制的数量
//
// _limit(3) => "LIMIT ?", []interface{}{3}
func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

// _where 生成 WHERE 语句
//
// 参数:
// values: 可变参数，第一个参数是条件描述，后面的参数是条件值
//
// 返回值:
// string: 生成的 WHERE 语句
// []interface{}: 条件值
//
// _where("Name = ?", "Tom") => "WHERE Name = ?", []interface{}{"Tom"}
func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %v", desc), vars
}

// _orderBy 生成 ORDER BY 语句
//
// 参数:
// values: 可变参数，第一个参数是排序字段
//
// 返回值:
// string: 生成的 ORDER BY 语句
// []interface{}: 空的参数列表
//
// _orderBy("Age") => "ORDER BY Age"
func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %v", values[0]), []interface{}{}
}

// _update 生成 UPDATE 语句
//
// 参数:
// values: 可变参数，第一个参数是表名，第二个参数是字段名和字段值的映射
//
// 返回值:
// string: 生成的 UPDATE 语句
// []interface{}: 字段值
//
// _update("users", map[string]interface{}{"Name": "Tom", "Age": 18}) => "UPDATE users SET Name = ?, Age = ?", []interface{}{"Tom", 18}
// 后面一般会跟 WHERE 子句，没有就是全改，不推荐全改
func _update(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for key, value := range m {
		keys = append(keys, key+" = ?")
		vars = append(vars, value)
	}
	return fmt.Sprintf("UPDATE %s SET %v", tableName, strings.Join(keys, ", ")), vars
}

// _delete 生成 DELETE 语句
//
// 参数:
// values: 可变参数，第一个参数是表名
//
// 返回值:
// string: 生成的 DELETE 语句
// []interface{}: 空的参数列表
//
// _delete("users") => "DELETE FROM users" 后续一般会跟 WHERE 子句，没有就是全删
func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

// _count 生成 COUNT 语句
//
// 参数:
// values: 可变参数，第一个参数是表名
//
// 返回值:
// string: 生成的 COUNT 语句
// []interface{}: 空的参数列表
//
// _count("users") => "SELECT count(*) FROM users"
func _count(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"count(*)"})
}
