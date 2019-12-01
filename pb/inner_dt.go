package pb

import (
	"github.com/Workiva/go-datastructures/set"
)

var dt_set *set.Set

func init() {
	dt_set = set.New()
}

// 添加内部数据定义
func addInnerDt(dt string) {
	dt_set.Add(dt)
}

// 检查是否存在字符串描述的数据定义
func isExistInnerDt(dt string) bool {
	return dt_set.Exists(dt)
}
