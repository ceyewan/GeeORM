// schema 用于实现对象（object）和表（table）之间的映射关系，是 ORM 的核心部分之一
package schema

import (
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

// Field 表示数据库表的一列
type Field struct {
	Name string // 列名
	Type string // 列的数据类型
	Tag  string // 列的额外信息（标签）
}

// Schema 表示数据库中的一张表
type Schema struct {
	Model      interface{}       // 表对应的对象
	Name       string            // 表名
	Fields     []*Field          // 表的所有列
	FieldNames []string          // 表的所有列名
	fieldMap   map[string]*Field // 列名到 Field 对象的映射
}

// GetField 根据列名获取 Field 对象
//
// 参数:
// name: 列名
//
// 返回值:
// *Field: 对应的 Field 对象
func (s *Schema) GetField(name string) *Field {
	return s.fieldMap[name]
}

// Parse 解析对象，创建 Schema
//
// 参数:
// dest: 要解析的对象
// d: 数据库方言
//
// 返回值:
// *Schema: 解析后的 Schema 对象
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		// 如果字段不是匿名字段且是导出字段，则创建 Field 对象
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			// 如果字段有 geeorm 标签，则使用标签的值
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

// RecordValues 返回对象中所有列的值
//
// 参数:
// dest: 要获取值的对象
//
// 返回值:
// []interface{}: 对象中所有列的值
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destvalue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destvalue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
