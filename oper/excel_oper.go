// 操作包，最终把excel转换成proto的核心包
// 我们把excel中的每一个sheel都会转换成以sheel名称命名的 .proto 文件，并且导出 .data的数据
// 每个sheel 都会在 .proto中会有一个以${sheelname}Storage结束的消息定义，Storage中只会放入
// 一个${sheelname}消息的数组
// 每一个sheel的前5行都是关于表的描术,具体打开一个xlsx就明白了
package oper

import (
	"fmt"
	"game_tools/excel"
	"game_tools/pb"
	"game_tools/urfave/cli.v2"
	"game_tools/util"
	"github.com/golang/protobuf/proto"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	excel_ext      = ".xlsx"   // excel文件扩展名
	data_ext       = ".data"   // 数据文件扩展名
	proto_ext      = ".proto"  // proto文件扩展名
	storage_suffix = "Storage" // 一个存储protobuf的仓库结构体后缀

	// 仓库结构中包含 Array 字段，有且只有一个这样的字段用于存储protobuf消息
	storage_field = "Array"
)

var (
	InputPath  string // 输入路径
	OutputPath string // 输出路径
	PackName   string // 定义到的包
)

// 恐慌捕捉
func panicCatch(err *error) {
	x := recover()
	if exitErr, ok := x.(cli.ExitCoder); ok {
		*err = exitErr
	} else if x != nil {
		panic(x)
	}
}

// 检查输入路径是否合法
func checkInputPath() {
	if InputPath == "" {
		panic(cli.Exit("xlsx的所在目录没有设置,请设置参数 -ip", 0))
	}
	InputPath, err := filepath.Abs(InputPath)
	if err != nil {
		panic(cli.Exit(err.Error(), 0))
	}

	InputPath = filepath.Clean(InputPath)
	fileInfo, err := os.Stat(InputPath)
	if os.IsNotExist(err) || !fileInfo.IsDir() {
		panic(cli.Exit("参数 -ip 不是一个有效目录, 请检查 -ip", 0))
	}
}

// 检查输出路径是否合法
func checkOutputPath() {
	if OutputPath == "" {
		panic(cli.Exit("输出路径没有设置，请设置参数 -op", 0))
	}
	OutputPath, err := filepath.Abs(OutputPath)
	if err != nil {
		panic(cli.Exit(err.Error(), 0))
	}

	OutputPath = filepath.Clean(OutputPath)
	fileInfo, err := os.Stat(OutputPath)
	if os.IsNotExist(err) || !fileInfo.IsDir() {
		panic(cli.Exit("参数 -op 不是一个有效目录, 请检查 -op", 0))
	}
}

// 检查包名是否合法
func checkPackName() {
	b, err := regexp.MatchString("[a-zA-Z]", PackName)
	if err != nil {
		panic(cli.Exit(err.Error(), 0))
	}
	if !b {
		panic(cli.Exit("包名只能是英文字母", 0))
	}
}

// 读excel并且写protobuf结构
func readExcelAndWriteProtoStruct(filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	xfile := excel.ReadStructWithFile(filename)
	pfile := pb.NewProto(xfile.Name)
	pfile.PackName = PackName

	fmt.Printf("[%s%s]开始进行构建.....\n", pfile.Name, proto_ext)
	for _, tabst := range xfile.Structs {
		msg := pb.NewMsgStruct(tabst.Name)
		for idx, field := range tabst.Fields {
			if field.IsSkip() {
				continue
			}
			msg.AddField(field.Type, field.Name, idx+1, false)
		}
		pfile.AddMsg(msg)
		fmt.Printf("\t结构:%s To 文件: %s%s\n", msg.Name, pfile.Name, proto_ext)

		//添加这个表的仓库
		msgStorage := pb.NewMsgStruct(msg.Name + storage_suffix)
		msgStorage.AddField(msg.Name, storage_field, 1, true)
		pfile.AddMsg(msgStorage)
	}
	pfile.WriteTo(OutputPath)
	fmt.Printf("[%s%s]已生成文件\n", pfile.Name, proto_ext)
}

// 字段赋值，错误返回字符串，否则返回空字符串
func field_assign(xfield *excel.Field, value *excel.Value, field reflect.Value) string {
	var finalval interface{}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value.Val)
	case reflect.Float64:
		floatval, err := strconv.ParseFloat(value.Val, 64)
		if err != nil {
			return err.Error() + "(转换float64错误)"
		}
		field.SetFloat(floatval)
		finalval = floatval
	case reflect.Float32:
		floatval, err := strconv.ParseFloat(value.Val, 32)
		if err != nil {
			return err.Error() + "(转换float32错误)"
		}
		field.SetFloat(floatval)
		finalval = floatval
	case reflect.Int32:
		intval, err := strconv.ParseInt(value.Val, 10, 32)
		if err != nil {
			return err.Error() + "(转换int32错误)"
		}
		field.SetInt(intval)
		finalval = intval
	case reflect.Int64:
		intval, err := strconv.ParseInt(value.Val, 10, 64)
		if err != nil {
			return err.Error() + "(转换int64错误)"
		}
		field.SetInt(intval)
		finalval = intval
	case reflect.Uint32:
		uintval, err := strconv.ParseUint(value.Val, 10, 32)
		if err != nil {
			return err.Error() + "(转换uint32错误)"
		}
		field.SetUint(uintval)
		finalval = uintval
	case reflect.Uint64:
		uintval, err := strconv.ParseUint(value.Val, 10, 64)
		if err != nil {
			return err.Error() + "(转换uint64错误)"
		}
		field.SetUint(uintval)
		finalval = uintval
	case reflect.Bool:
		boolval, err := strconv.ParseBool(value.Val)
		if err != nil {
			return err.Error() + "(转换bool错误)"
		}
		field.SetBool(boolval)
		finalval = boolval
	}
	if xfield.Range != nil {
		if !xfield.Range.Check(finalval) {
			return "不符合范围"
		}
	}
	return ""
}

// 检查数据的合法性,返回bool,string, 当bool=true:说明字段跳过，string有值说明检查有问题，描述的了具体问题
func check_data_rightful(xfield *excel.Field, value *excel.Value, protoVal reflect.Value) (bool, string) {
	if value.Name != xfield.Name {
		return false, "行数据中表达的字段与结构描术的字段不一致,代码有BUG,请检查"
	}

	fieldInfo, ok := protoVal.Type().FieldByName(value.Name)
	if !ok {
		return false, "字段在proto结构中不存在.."
	}
	//字段不是protobuf数据
	if _, ok := fieldInfo.Tag.Lookup("protobuf"); !ok {
		return true, ""
	}
	return false, ""
}

// 构建一个指定结构的仓库存储消息
func newStorageMsg(st *excel.TableStruct, protoMsgArr []proto.Message) (proto.Message, error) {
	//构建出Storage来，每一种结构都会有一个Storage,Storage存放的就是这个结构的数组
	storageFullName := fmt.Sprintf("%s.%s%s", PackName, st.Name, storage_suffix)
	storage_type := proto.MessageType(storageFullName)
	if storage_type == nil {
		return nil, cli.Exit(fmt.Sprintf("%s类型不存，请检查proto描述文件是否编译", storageFullName), 0)
	}
	storageMsg := reflect.New(storage_type.Elem()).Interface()
	storageVal := reflect.ValueOf(storageMsg).Elem()
	storageField := storageVal.FieldByName(storage_field)

	var arrayVal reflect.Value
	for _, msg := range protoMsgArr {
		arrayVal = reflect.Append(storageField, reflect.ValueOf(msg))
		storageField.Set(arrayVal)
	}
	return storageMsg.(proto.Message), nil
}

// 检查是否为空行，如果数据都为空字符串，则表示为空行
func isEmptyRow(row excel.ROW) bool {
	empty := true
	for _, value := range row {
		if value.Val != "" {
			empty = false
			break
		}
	}
	return empty
}

// 构建一个指定结构的原始protobuf消息结构
func newOriginProtoMsg(st *excel.TableStruct) (proto.Message, string) {
	fullName := fmt.Sprintf("%s.%s", PackName, st.Name)
	st_type := proto.MessageType(fullName)
	if st_type == nil {
		return nil, "protobuf结构的类型不存在，可能是没编译进去"
	}
	msg := reflect.New(st_type.Elem()).Interface()
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, "转换成protobuf失败"
	}
	return protoMsg, ""
}

// 写数据到文件
func writeDataToFile(st *excel.TableStruct, storageMsg proto.Message) error {
	bytearr, err := proto.Marshal(storageMsg)
	if err != nil {
		return cli.Exit(
			fmt.Sprintf("[%s]protobuf序列化时错误，ERR=%s", st.Name, err.Error()),
			0,
		)
	}
	filename := OutputPath + string(os.PathSeparator) + st.Name + data_ext
	err = util.WriteFileWithBytes(filename, bytearr)
	if err != nil {
		return cli.Exit(fmt.Sprintf("[%s]数据文件保存失败，ERR=%s", filename, err.Error()),
			0,
		)
	}

	fmt.Printf("[%s]数据文件保存成功\n", filename)
	return nil
}

// 读excel并写proto数据
func readExcelAndWriteProtoData(filename string, wg *sync.WaitGroup) error {
	fmt.Printf("写数据文件 %s\n", filename)
	defer wg.Done()

	errformat := "[file=%s,st=%s,ridx=%d,idx=%d,v.n=%s,f.n=%s]%s"

	xfile := excel.ReadDataWithFile(filename)
	for _, st := range xfile.Structs {
		originMsg, errorinfo := newOriginProtoMsg(st)
		if errorinfo != "" {
			return cli.Exit(fmt.Sprintf(errformat, xfile.Name, st.Name, -1, -1, "", "",
				errorinfo), 0)
		}
		rows := xfile.ValueMap[st.Name]
		var protoMsgArr []proto.Message

		for ridx, row := range rows {
			if isEmptyRow(row) {
				//如果是空行则跳过
				continue
			}
			protoMsg := proto.Clone(originMsg)
			protoVal := reflect.ValueOf(protoMsg).Elem()

			for idx, xfield := range st.Fields {
				if xfield.IsSkip() {
					continue
				}
				value := row[idx]

				//检查数据的合法性
				ok, errinfo := check_data_rightful(xfield, value, protoVal)
				if errinfo != "" {
					return cli.Exit(fmt.Sprintf(errformat, xfile.Name, st.Name, ridx, idx, value.Name, xfield.Name,
						errinfo), 0)
				}
				//字段跳过
				if ok {
					continue
				}

				field := protoVal.FieldByName(value.Name)
				errinfo = field_assign(xfield, value, field)

				if errinfo != "" {
					return cli.Exit(fmt.Sprintf(errformat, xfile.Name, st.Name, ridx, idx, value.Name, xfield.Name,
						errinfo), 0)
				}
			}
			protoMsgArr = append(protoMsgArr, protoMsg)
		}

		if len(protoMsgArr) > 0 {
			storageMsg, err := newStorageMsg(st, protoMsgArr)
			if err != nil {
				return err
			}

			err = writeDataToFile(st, storageMsg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// 返回命令行参数定义
func Params() []cli.Flag {
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:        "inputpath",
			Aliases:     []string{"ip"},
			Usage:       "xlsx文件所在路径",
			Destination: &InputPath,
		},

		&cli.PathFlag{
			Name:        "outputpath",
			Aliases:     []string{"op"},
			Usage:       "最后结果输出路径",
			Destination: &OutputPath,
		},

		&cli.StringFlag{
			Name:        "packname",
			Aliases:     []string{"pn"},
			Usage:       "包名",
			Destination: &PackName,
		},
	}
	return flags
}

// excel结构导出成.proto文件
func ExcelStructToProto(ctx *cli.Context) (err error) {
	defer panicCatch(&err)
	checkInputPath()
	checkOutputPath()
	checkPackName()

	deleteAllProtoFile(InputPath)
	files, err := scanAllExcelFile(InputPath)
	if err != nil {
		return cli.Exit(err.Error(), 0)
	}

	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, filename := range files {
		go readExcelAndWriteProtoStruct(filename, &wg)
	}
	wg.Wait()
	return nil
}

// 删除所有.data文件
func deleteAllDataFile(dir string) error {
	err := util.DelFilesWithDir(dir, func(file os.FileInfo) bool {
		if strings.ToLower(filepath.Ext(file.Name())) == data_ext {
			return true
		}
		return false
	})
	return err
}

// 删除所有 .proto 文件
func deleteAllProtoFile(dir string) error {
	err := util.DelFilesWithDir(dir, func(file os.FileInfo) bool {
		if strings.ToLower(filepath.Ext(file.Name())) == proto_ext {
			return true
		}
		return false
	})
	return err
}

// 扫描所有excel文件
func scanAllExcelFile(dir string) ([]string, error) {
	files, err := util.ListFiles(dir, func(file os.FileInfo) bool {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == excel_ext {
				return true
			}
		}
		return false
	})
	return files, err
}

// 导出数据到Proto消息中去，先一定是生成了proto消息，具体可以参考工具的批处理
// 这里肯定是先导出了proto消息结构体，然后重新编译了工具，才能调用
func ExcelDataToProto(ctx *cli.Context) (err error) {
	defer panicCatch(&err)
	checkInputPath()
	checkOutputPath()
	checkPackName()

	// 删除掉所有数据文件
	err = deleteAllDataFile(OutputPath)

	if err != nil {
		return cli.Exit("[ERR="+err.Error()+"]数据文件删除失败", 0)
	}

	files, err := scanAllExcelFile(InputPath)
	if err != nil {
		return cli.Exit("[ERR="+err.Error()+"]获取excel文件失败", 0)
	}

	filenum := len(files)

	var wg sync.WaitGroup
	wg.Add(filenum)

	errchan := make(chan error, filenum)

	for i := 0; i < filenum; i++ {
		// 下面的协程会让传入到方法里i的值不可预测，这里进行一次i的copy，只有4个字节
		// 如果用 range 那么对于传入的字符串也是不可预测的，同理我们可以通过进行一次
		// copy来避免，但字符串的copy远没有int的copy开销小，所以这里改成了for ;; 的形式
		_i := i
		go func() {
			e := readExcelAndWriteProtoData(files[_i], &wg)
			errchan <- e
		}()
	}

	//for _, filename := range files {
	//	_filename := filename
	//	go func() {
	//		e := readExcelAndWriteProtoData(_filename, &wg)
	//		errchan <- e
	//	}()
	//
	//	fmt.Println("xxx")
	//}

	var resultErr error

	for i := 0; i < filenum; i++ {
		select {
		case e := <-errchan:
			if e != nil {
				resultErr = e
				break
			}
		}
	}
	wg.Wait()

	if resultErr != nil {
		deleteAllDataFile(OutputPath)
	}
	return resultErr
}
