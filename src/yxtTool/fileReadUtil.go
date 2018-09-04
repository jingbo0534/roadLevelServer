package yxtTool

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var keys map[string]string = make(map[string]string, 1024)

/**
读取配置文件，配置文件地址
src/static/prop.txt
*/
func ReadProperty() map[string]string {
	return readFile("prop.txt")
}

/**
读文件方法 2 按行读
*/
func ScannerRead(path string) map[string]string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if file == nil {
		fmt.Println("文件不存在!")
		return nil
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		str = strings.TrimSpace(str)
		if str != "" {
			if str[0:1] != "#" {
				ss := strings.Split(str, " ")
				keys[ss[0]] = ss[1]
			}
		}
	}
	return keys
}

/**
读文件方法 1
*/
func readFile(path string) map[string]string {

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0766)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if file == nil {
		fmt.Println("文件不存在!")
		return nil
	}
	b := make([]byte, 1024)
	count, err := file.Read(b)
	b = b[0:count]
	str := string(b)
	strs := strings.Split(str, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if str != "" {
			if str[0:1] != "#" {
				ss := strings.Split(str, " ")
				if len(ss) == 2 {
					keys[ss[0]] = ss[1]
				}
			}
		}
	}
	return keys
}
