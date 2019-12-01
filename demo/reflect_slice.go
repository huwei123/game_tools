// 下面是关于如何在用反射的方法，在一个结构体中设置一个切片字段的值，分别用了3种方法
// 此包不用关心，用于测试
package demo

import (
	"reflect"
)

type test struct {
	Age int
}

type T struct {
	Name     string
	Children []*test
}

func AppendSlice(i, e interface{}) {
	v := reflect.ValueOf(i).Elem()
	v = v.FieldByName("Children")

	arr1 := reflect.Append(v, reflect.ValueOf(e))
	v.Set(arr1)
}

func main11() {
	t := &T{}
	s := reflect.ValueOf(t).Elem()
	s.Field(0).SetString("ABC")

	test1 := &test{
		Age: 1,
	}

	test2 := &test{
		Age: 2,
	}

	AppendSlice(t, test1)
	AppendSlice(t, test2)
}

func main22() {
	t := &T{}
	s := reflect.ValueOf(t).Elem()
	s.Field(0).SetString("ABC")

	test1 := &test{
		Age: 1,
	}

	test2 := &test{
		Age: 2,
	}

	sliceValue := reflect.ValueOf([]*test{
		test1, test2,
	})

	a0 := s.FieldByName("Children")
	a0.Set(sliceValue)
}

func main33() {
	t := &T{}
	s := reflect.ValueOf(t).Elem()
	s.Field(0).SetString("ABC")

	test1 := &test{
		Age: 1,
	}

	test2 := &test{
		Age: 2,
	}

	a0 := s.FieldByName("Children")

	e0 := make([]reflect.Value, 0)
	e0 = append(e0, reflect.ValueOf(test1))
	e0 = append(e0, reflect.ValueOf(test2))

	val_arr1 := reflect.Append(a0, e0...)
	a0.Set(val_arr1)
}
