// 进行excel 读取操作的包，把符合定义的excel文件读取成后续程序能操作的数据结构
package excel

import (
	"errors"
	"fmt"
	"game_tools/dt"
	"game_tools/util"
	"github.com/tealeg/xlsx"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ext_name      = ".xlsx" //excel的扩展名
	table_symbol  = "@"     //表符号
	struct_top    = 5       //表示结构的行是前多少行
	skip_field    = "null"  //需要跳过的字段
	primary_field = "key"   //主键字段(标有主键字段的会检查数据的唯一性)
)

// 下面这些常量表示的行，一定是有内容的，后面的数据是根据这些来解析的
const (
	explain_ln   = iota // 说明行
	attribute_ln        // 属性行
	type_ln             // 类型行
	range_ln            // 范围行
	field_ln            // 字段行
)

// 范围(范围一定是有符号整型，即便是类型是浮点数，也需要用整型表示范围)
type Range struct {
	Min int64 // 最小值
	Max int64 // 最大值
}

// 字段(在excel中无论表示什么字段，都会原样添加进来，即便标记为跳过的字段)
type Field struct {
	Name  string //字段名称
	Attr  string //字段属性
	Type  string //字段类型
	Desc  string //字段描述
	Range *Range //字段范围
}

// 值(在excel中读取到的值，并记录下这个值属于哪个字段)
type Value struct {
	Name string //列名，对应字段名
	Val  string //值
}

// 一行由多个值组成
type ROW []*Value

//表结构
type TableStruct struct {
	Name     string            //表名称(或称为结构名称)
	Fields   []*Field          //表结构中的所有字段
	FieldMap map[string]*Field //为了方便按名称查找对应的映射字段
}

// 文件 (描述一个excel文件，这份描述已经把excel按规则进行了结构化)
type File struct {
	Name      string                  // 文件名
	Structs   []*TableStruct          // 结构数组(一个excel中可能有多张表)
	StructMap map[string]*TableStruct // 文件中所有表结构的映射
	ValueMap  map[string][]ROW        // 文件中表里的所有数据
}

// 是否为跳过字段
func (f *Field) IsSkip() bool {
	if f.Attr == skip_field {
		return true
	}
	return false
}

// 是否为主键字段(目前暂未应用，后续需要加入)
func (f *Field) IsPrimaryKey() bool {
	if f.Attr == primary_field {
		return true
	}
	return false
}

// 检查值是否符合范围
func (r *Range) Check(val interface{}) bool {
	switch v := val.(type) {
	case int32:
		if int64(v) >= r.Min && int64(v) <= r.Max {
			return true
		}
	case int64:
		if v >= r.Min && v <= r.Max {
			return true
		}
	case uint32:
		if int64(v) >= r.Min && int64(v) <= r.Max {
			return true
		}
	case uint64:
		if v >= uint64(r.Min) && v <= uint64(r.Max) {
			return true
		}
	case float32:
		if v >= float32(r.Min) && v <= float32(r.Max) {
			return true
		}
	case float64:
		if v >= float64(r.Min) && v <= float64(r.Max) {
			return true
		}
	}
	return false
}

// 是否为xlsx文件
func isXlsx(filename string) bool {
	if filepath.Ext(filename) == ext_name {
		return true
	}
	return false
}

// 寻找到xlsx文件中表示为表的sheet
func findTable(xfile *xlsx.File) []*xlsx.Sheet {
	var sheets []*xlsx.Sheet
	if len(xfile.Sheets) > 0 {
		for _, sheet := range xfile.Sheets {
			if strings.HasPrefix(sheet.Name, table_symbol) {
				sheets = append(sheets, sheet)
			}
		}
	}
	return sheets
}

// 从全路径的文件名中分析出表的前缀名
func prefixTableName(filename string) string {
	name := filepath.Base(filename)
	name = name[0:strings.IndexRune(name, '.')]
	return name
}

// 读数据结果，参数num=0:表示读全部的行，>0:表示读前num行
func read_data_result(sheet *xlsx.Sheet, num int) ([][]string, error) {
	output := [][]string{}
	for idx, row := range sheet.Rows {
		if num > 0 && idx >= num {
			break
		}
		if row == nil {
			continue
		}
		r := []string{}
		for _, cell := range row.Cells {
			str, err := cell.FormattedValue()
			if err != nil {
				if numErr, ok := err.(*strconv.NumError); ok && numErr.Num == "" {
					str = ""
				} else {
					return output, err
				}
			}
			r = append(r, str)
		}
		output = append(output, r)
	}
	return output, nil
}

// 解析范围类型
func parse_range(str string) *Range {
	if str == "" || str == "0" || !strings.ContainsRune(str, '~') {
		return nil
	}
	strArr := strings.Split(str, "~")
	if len(strArr) != 2 {
		return nil
	}
	min, err1 := strconv.Atoi(strArr[0])
	max, err2 := strconv.Atoi(strArr[1])
	if err1 != nil || err2 != nil {
		return nil
	}
	return &Range{
		Min: int64(min),
		Max: int64(max),
	}
}

// 读取数据
func read_data(file *File, sheet *xlsx.Sheet) error {
	//读数据前先把结构读出来
	err := read_struct(file, sheet)
	if err != nil {
		return err
	}
	//最后一个结构就是当前的结构
	st := file.Structs[len(file.Structs)-1]
	data, err := read_data_result(sheet, 0) //读出所有数据，包括前面的结构数据
	if err != nil {
		return err
	}

	var rows []ROW
	for ridx, rowdata := range data {
		if ridx < struct_top {
			continue
		}

		rdlen := len(rowdata) //行数据的长度

		var row ROW
		for fidx, field := range st.Fields {
			val := &Value{
				Name: field.Name,
				Val: func() string {
					if fidx < rdlen {
						return rowdata[fidx]
					} else {
						return ""
					}
				}(),
			}
			row = append(row, val)
		}
		rows = append(rows, row)
	}

	if file.ValueMap == nil {
		file.ValueMap = make(map[string][]ROW)
	}
	file.ValueMap[st.Name] = rows
	return nil
}

// 读取结构
func read_struct(file *File, sheet *xlsx.Sheet) error {
	st := &TableStruct{}
	st.Name = fmt.Sprintf("%s_%s",
		strings.ToLower(file.Name),
		strings.ToLower(sheet.Name[1:]),
	) //结构名称(全部转成小写)
	st.Name = util.CamelString(st.Name) //一次性转成驼峰格式

	output, err := read_data_result(sheet, struct_top)
	if err != nil {
		return err
	}
	if len(output) < struct_top {
		return errors.New("结构行不能少于5行，否则定义是不全的")
	}
	if len(output[explain_ln]) <= 0 || len(output[attribute_ln]) <= 0 ||
		len(output[type_ln]) <= 0 || len(output[range_ln]) <= 0 || len(output[field_ln]) <= 0 {
		return errors.New("结构列不能小于0，否则定义是不全的")
	}

	len := len(output[field_ln])
	for i := 0; i < len; i++ {
		val := output[field_ln][i]
		//如果从字段行中拿到的是空字符串，则跳过字段构建,
		if val == "" {
			continue
		}
		//if strings.ToLower(output[attribute_ln][i]) == skip_field {
		//	continue
		//}
		field := &Field{}
		field.Name = util.CamelString(strings.ToLower(val))
		field.Attr = strings.ToLower(output[attribute_ln][i])
		field.Desc = output[explain_ln][i]
		field.Type = strings.ToLower(output[type_ln][i])

		if !dt.Check(field.Type) {
			return errors.New(fmt.Sprintf("[file:%s,tab:%s,field:%s,type:%s]类型定义错误",
				file.Name, sheet.Name, field.Name, field.Type))
		}

		field.Range = parse_range(output[range_ln][i])
		st.Fields = append(st.Fields, field)
		if st.FieldMap == nil {
			st.FieldMap = make(map[string]*Field)
		}
		_, ok := st.FieldMap[field.Name]
		if ok {
			return errors.New(fmt.Sprintf("[file:%s,tab:%s,field:%s]字段重复定义",
				file.Name, sheet.Name, field.Name))
		}
		st.FieldMap[field.Name] = field
	}

	file.Structs = append(file.Structs, st)
	if file.StructMap == nil {
		file.StructMap = make(map[string]*TableStruct)
	}
	_, ok := file.StructMap[st.Name]
	if ok {
		return errors.New(fmt.Sprintf("[file:%s,tab:%s]表重复定义", file.Name, sheet.Name))
	}
	file.StructMap[st.Name] = st
	return nil
}

// 内部读取方法，外界是调用不到的，对于excel的读取分为读结构和读数据，这二种操作的本质是一样的
// 只是对于解析的方式不同，所以统一都调read_file，内部传不同的读取方法，如: read_data,read_struct
func read_file(filename string, fn func(*File, *xlsx.Sheet) error) (*File, error) {
	filename = filepath.Clean(filename)
	//先检查是否为xlsx文件
	if !isXlsx(filename) {
		return nil, errors.New(fmt.Sprintf("这不是一个xlsx文件，file: %s", filename))
	}
	xfile, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	sheets := findTable(xfile)
	file := &File{
		Name: prefixTableName(filename),
	}
	for _, sheet := range sheets {
		err = fn(file, sheet)
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}

// 通过文件读取结构
func ReadStructWithFile(filename string) *File {
	file, err := read_file(filename, read_struct)
	if err != nil {
		panic(err)
	}
	return file
}

// 通过文件读取数据
func ReadDataWithFile(filename string) *File {
	file, err := read_file(filename, read_data)
	if err != nil {
		panic(err)
	}
	return file
}

////通过目录读取结构
//func ReadStructWithDir(dir string) ([]TableStruct, error) {
//	//todo
//	return nil,nil
//}
