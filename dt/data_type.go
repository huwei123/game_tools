/*
 关于proto数据类型定义,主要用于在后面操作中进行类型效验,把所有在excel中定义会使用的数据类型都在这里描述
*/
package dt

import "reflect"

// 数据内型定义，用字符串描述
type dataType struct {
	Double   string
	Float    string
	Int32    string
	Int64    string
	Uint32   string
	Uint64   string
	Sint32   string
	Sint64   string
	Fixed32  string
	Fixed64  string
	Sfixed32 string
	Sfixed64 string
	Bool     string
	String   string
}

var (
	define *dataType     // 数据类型定义变量(init()方法中会进行初始化)
	v      reflect.Value // 之后方法中用于反射define变量的(init()方法中会进行初始化)
)

// 初始化 define 与 v 二个变量，后期的检查主要是利用runtime的反射功能对
// v进行查找
func init() {
	define = &dataType{
		Double:   "double",
		Float:    "float",
		Int32:    "int32",
		Int64:    "int64",
		Uint32:   "uint32",
		Uint64:   "uint64",
		Sint32:   "sint32",
		Sint64:   "sint64",
		Fixed32:  "fixed32",
		Fixed64:  "fixed64",
		Sfixed32: "sfixed32",
		Sfixed64: "sfixed64",
		Bool:     "bool",
		String:   "string",
	}

	v = reflect.ValueOf(define)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
}

// 检查字符串参数val所表示的数据类型是否存在定义
func Check(val string) bool {
	if val == "" {
		return false
	}
	size := v.NumField()
	for i := 0; i < size; i++ {
		target := v.Field(i).String()
		if val == target {
			return true
		}
	}
	return false
}
