package util

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 列出指定目录下的所有文件
// 会通过 filterFn 传入的方法进行过滤
func ListFiles(dir string, filterFn func(file os.FileInfo) bool) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	rc_files := make([]string, 0, 10)

	for _, file := range files {
		if file.IsDir() {
			tmp_files, err := ListFiles(dir+string(filepath.Separator)+file.Name(), filterFn)
			if err != nil {
				return nil, err
			} else {
				rc_files = append(rc_files, tmp_files...)
			}
		} else {
			if filterFn != nil {
				if filterFn(file) {
					rc_files = append(rc_files, dir+string(filepath.Separator)+file.Name())
				}
			} else {
				rc_files = append(rc_files, dir+string(filepath.Separator)+file.Name())
			}
		}
	}
	return rc_files, nil
}

// 删除传入的所有文件
func DelFiles(files ...string) {
	for _, file := range files {
		os.Remove(file)
	}
}

// 删除目录下的所有文件
// 通过 filterFn 传入的方法进行过滤
func DelFilesWithDir(dir string, filterFn func(file os.FileInfo) bool) error {
	files, err := ListFiles(dir, filterFn)
	if err != nil {
		return err
	}
	DelFiles(files...)
	return nil
}

// 写入内容到指定的文件
// fileName: 指定的文件
// content: 写入的内容
func WriteFileWithString(fileName string, content string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	} else {
		defer f.Close()
		_, err := f.WriteString(content)
		return err
	}
}

// 写入内容到指定的文件
func WriteFileWithBytes(fileName string, content []byte) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	} else {
		defer f.Close()
		_, err := f.Write(content)
		return err
	}
}

// 读取文件中的第一行，返回[]string，每一行为一个下标
func ReadFileWithLines(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		lines = append(lines, line)
	}
	return lines, nil
}
