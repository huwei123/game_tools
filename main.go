package main

import (
	"fmt"
	"game_tools/oper"
	"game_tools/urfave/cli.v2"
	"game_tools/util"
	"os"
	"strings"
	//#
)

var modifFileName string
var modifValue string

func main() {

	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:    "excel2struct",
				Aliases: []string{"struct"},
				Usage:   "excel表导出proto结构的描述文件",
				Flags:   oper.Params(),
				Action:  oper.ExcelStructToProto,
			},

			{
				Name:    "excel2data",
				Aliases: []string{"data"},
				Usage:   "excel表导出proto数据",
				Flags:   oper.Params(),
				Action:  oper.ExcelDataToProto,
			},

			{
				Name:  "modif",
				Usage: "用于修文main.go的导出项",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "modifyfile",
						Aliases:     []string{"mf"},
						Usage:       "需要要修改的文件",
						Destination: &modifFileName,
					},
					&cli.StringFlag{
						Name:        "modifyvalue",
						Aliases:     []string{"mv"},
						Usage:       "替换的值",
						Destination: &modifValue,
					},
				},
				Action: modif,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("执行出错 Err: %v\n", err)
	} else {
		fmt.Println("执行完成....")
	}
}

func modif(ctx *cli.Context) (err error) {
	lines, err := util.ReadFileWithLines(modifFileName)

	modifyIdx := -1
	for idx, line := range lines {
		// 找到修改行
		if strings.Contains(line, "//#") {
			modifyIdx = idx
			break
		}
	}

	if modifyIdx >= 0 {
		fmt.Printf("替换成功[%s][%s]\n", lines[modifyIdx], modifValue)
		lines[modifyIdx] = modifValue + "\n"
		if !strings.Contains(lines[len(lines)-1], "}") {
			lines = append(lines, "}\n")
		}

		var sb strings.Builder

		for _, line := range lines {
			sb.WriteString(line)
		}

		util.WriteFileWithString(modifFileName, sb.String())
	}

	return nil
}
