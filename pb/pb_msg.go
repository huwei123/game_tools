package pb

import (
	"fmt"
	"game_tools/dt"
	"strings"
)

// proto文件中关于消息结构的定义
type MsgStruct struct {
	Name   string  // 消息名称
	Fields []Field // 消息中的字段
}

// 代表消息中的一个字段
type Field struct {
	T       string // 字段类型
	N       string // 字段名称
	S       int    // 字段序列
	IsArray bool   // 是否数组
}

// 生成消息中的字段行
func (l *Field) Line() string {
	var format string
	if l.IsArray {
		format = "\trepeated %s  %s  = %d;\n"
	} else {
		format = "\t%s  %s  = %d;\n"
	}
	str := fmt.Sprintf(format, l.T, l.N, l.S)
	return str
}

// 定义消息的开始行
type msgStartLine string //消息开始行

// 定义消息的结束行
type msgEndLine struct{}

// 生成消息的开始行
func (l msgStartLine) Line() string {
	str := fmt.Sprintf("message %s {\n", string(l))
	return str
}

// 生成消息的结束行
func (l msgEndLine) Line() string {
	str := "}\n"
	return str
}

////先项
//type Option struct {
//	Key    string
//	Val    string
//}

// 构建一个proto的消息结构
func NewMsgStruct(name string) *MsgStruct {
	msg := &MsgStruct{
		Name: name,
	}
	return msg
}

// 添加字段 isArray=true：说明是数组字段, isArray=false: 不是数组字段
func (m *MsgStruct) AddField(t string, n string, s int, isArray bool) *MsgStruct {
	field := Field{T: t, N: n, S: s, IsArray: isArray}
	if !dt.Check(t) && !isExistInnerDt(t) {
		panic(fmt.Sprintf("[%v]未知类型", field))
	}
	m.Fields = append(m.Fields, field)
	return m
}

// 生成消息在proto中的字符串
func (m *MsgStruct) ToString() string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString(msgStartLine(m.Name).Line()) //消息开始行

	if len(m.Fields) > 0 {
		for _, field := range m.Fields {
			stringBuilder.WriteString(field.Line())
		}
	}

	stringBuilder.WriteString(msgEndLine{}.Line())
	return stringBuilder.String()
}
