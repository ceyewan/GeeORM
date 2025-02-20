package session

import (
	"geeorm/log"
	"reflect"
)

// Hooks 常量
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

// CallMethod 方法用于调用钩子函数
//
// CallMethod 根据提供的方法名和可选的值来调用钩子函数。
// 如果提供了值，则从该值中查找方法，否则从 Session 的模型中查找方法。
// 调用方法时，会传递当前的 Session 作为参数。
// 如果方法调用返回错误，则会记录错误日志。
//
// 参数:
//   - method: 要调用的方法名。
//   - value: 可选的值，如果提供了该值，则从该值中查找方法。
func (s *Session) CallMethod(method string, value interface{}) {
	// 从 Session 的 Model 中查找方法
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		// 从提供的值中查找方法
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	// 准备调用函数的参数，这里传递的是当前的 Session
	param := []reflect.Value{reflect.ValueOf(s)}
	if fm.IsValid() {
		// 调用方法并获取返回值
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
	return
}
