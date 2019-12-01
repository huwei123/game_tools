package pb

import (
	"fmt"
	"game_tools/util"
	"path/filepath"
	"strings"
)

const (
	Syntax2 = 2 //proto2
	Syntax3 = 3 //proto3
)

// 表示文件中的一行(proto描术文件都是一行一行的，这里用于抽像每一行)
type FileLine interface {
	Line() string
}

// 选项 (proto描术文件中选项的描述)
type Option struct {
	Key string
	Val string
}

// 生成选项行
func (l *Option) Line() string {
	str := fmt.Sprintf("option %s = \"%s\";\n", l.Key, l.Val)
	return str
}

// 重新定义类型为了实现FileLine接口，代表文件中的一行
type syntaxLine int  // 语法行
type packLine string // 包名行

type newLine struct{} // 换行

// 生成语法行
func (l syntaxLine) Line() string {
	str := fmt.Sprintf("syntax = \"proto%d\";\n", int(l))
	return str
}

// 生成包名行
func (l packLine) Line() string {
	str := fmt.Sprintf("package %s;\n", string(l))
	return str
}

// 换行
func (l newLine) Line() string {
	return "\n"
}

// proto文件的描术对构
type ProtoFile struct {
	Name     string       // 文件名
	SyntaxLv int          // 语法等级
	PackName string       // 包名
	Options  []Option     // 选项
	Messages []*MsgStruct // 消息结构
}

// 构建一个name名称的proto文件
func NewProto(name string) *ProtoFile {
	return &ProtoFile{
		Name:     name,
		SyntaxLv: Syntax3,
	}
}

// 向proto文件中添加选项
func (this *ProtoFile) AddOption(key string, value string) *ProtoFile {
	opt := Option{
		Key: key,
		Val: value,
	}
	this.Options = append(this.Options, opt)
	return this
}

// 向proto文件中添加消息
func (this *ProtoFile) AddMsg(msgStruct *MsgStruct) {
	this.Messages = append(this.Messages, msgStruct)
	addInnerDt(msgStruct.Name)
}

// 把ProtoFile生成的字符串写到指定的目录中去
func (this *ProtoFile) WriteTo(dir string) {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(syntaxLine(this.SyntaxLv).Line())
	stringBuilder.WriteString(packLine(this.PackName).Line())

	//选项
	if len(this.Options) > 0 {
		for _, opt := range this.Options {
			stringBuilder.WriteString(opt.Line())
		}
	}
	stringBuilder.WriteString(newLine{}.Line())

	if len(this.Messages) > 0 {
		for _, msg := range this.Messages {
			stringBuilder.WriteString(msg.ToString())
			stringBuilder.WriteString(newLine{}.Line())
		}
	}

	dir = filepath.Clean(dir)
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(fmt.Sprintf("路径格式化中出错....dir:%s", dir))
	}
	fileName := fmt.Sprintf("%s%c%s.proto", dir, filepath.Separator, this.Name)
	writeToFile(fileName, stringBuilder.String())
}

// 写文件
func writeToFile(fileName string, content string) {
	err := util.WriteFileWithString(fileName, content)
	if err != nil {
		panic(fmt.Sprintf("生成 %s 失败 err=%v", fileName, err))
	}
}
