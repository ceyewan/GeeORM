package session

import (
	"errors"
	"geeorm/clause"
	"reflect"
)

// Insert 插入记录到数据库中
//
// 参数:
// values: 要插入的记录
//
// 返回值:
// int64: 受影响的行数
// error: 如果插入过程中发生错误，返回错误信息
//
// 示例:
// User1 := &User{"Tom", 18}、User2 := &User{"Sam", 25}
// affected, err := s.Insert(User1, User2)
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		// tables.Name 是 User，tables.FieldNames 是 [Name, Age]
		tables := s.Model(value).RefTable()
		// INSERT INTO $tableName ($fields)
		s.clause.Set(clause.INSERT, tables.Name, tables.FieldNames)
		recordValues = append(recordValues, tables.RecordValues(value))
	}
	// VALUES (?, ?), (?, ?)
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	// INSERT INTO $tableName ($fields) VALUES (?, ?), (?, ?)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Find 查找记录并填充到 values 中
//
// 参数:
// values: 要填充的记录
//
// 返回值:
// error: 如果查找过程中发生错误，返回错误信息
//
// 示例:
// var users []User
// err := s.Find(&users)
func (s *Session) Find(values interface{}) error {
	// 将 values 转换为 reflect.Value 类型并获取其指针指向的值，即 []User
	destValue := reflect.Indirect(reflect.ValueOf(values))
	// 获取 []User 中的元素类型 User，即 destType 是 User 类型
	destType := destValue.Type().Elem()
	// 获取 User 对应的表结构
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	// SELECT $fields FROM $tableName，即 SELECT Name, Age FROM users
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	// 执行代码
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	// 遍历查询结果并将结果填充到 values 中
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		destValue.Set(reflect.Append(destValue, dest))
	}
	return rows.Close()
}

// Update 更新记录
//
// 参数:
// kv: 要更新的键值对
//
// 返回值:
// int64: 受影响的行数
// error: 如果更新过程中发生错误，返回错误信息
//
// 示例:
// affected, err := s.Update("Age", 30)
// affected, err := s.Update(map[string]interface{}{"Age": 30, "Name": "Tom"})
func (s *Session) Update(kv ...interface{}) (int64, error) {
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete 删除记录
//
// 返回值:
// int64: 受影响的行数
func (s *Session) Delete() (int64, error) {
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Count 返回记录总数
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Where 添加 WHERE 子句，返回值是 *Session 可以链式调用
func (s *Session) Where(desc string, args ...interface{}) *Session {
	s.clause.Set(clause.WHERE, append([]interface{}{desc}, args...)...)
	return s
}

// Limit 添加 LIMIT 子句，返回值是 *Session 可以链式调用
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// OrderBy 添加 ORDER BY 子句，返回值是 *Session 可以链式调用
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}
