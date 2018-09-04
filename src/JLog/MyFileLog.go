package JLog

import (
	"fmt"
	"log"
	"os"
	"static"
	"time"
	"yxtTool"
)

const PRINT = false


func WriteClientMsg(logMsg string, dirName string, fileName string) {

	curTimeStr := time.Now().Format(static.TimeLayOut3)
	dirName = yxtTool.MergeString("roadLevel/", dirName)
	ClientDataFilePath := yxtTool.MergeString(dirName, "/"+fileName, curTimeStr, ".txt")
	out, err := os.OpenFile(ClientDataFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	defer out.Close()
	if err != nil {
		err = os.MkdirAll(dirName, 0777)
		if err != nil {
			//创建目录失败
			log.Println("创建目录失败")
		} else {
			out, err = os.OpenFile(ClientDataFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		}
	}
	offSet, err := out.Seek(0, 2)
	if err != nil {
		offSet = 0
	}
	logMsg = yxtTool.MergeString(time.Now().Format(static.TimeLayOut1), "", logMsg, "\r\n")
	out.WriteAt([]byte(logMsg), offSet)
}

func WriteClientMsgWithPrint(logMsg string, dirName string, fileName string, bool bool) {
	if bool {
		fmt.Println(yxtTool.MergeString(dirName, ";", logMsg))
	}
	WriteClientMsg(logMsg, dirName, fileName)
}

func PrintSqlError(logMsg string) {
	WriteClientMsg(logMsg, "sqlErr", "sqlInfo")
	log.Println(logMsg)
}
func PrintSysErr(logMsg string) {
	WriteClientMsg(logMsg, "sysErr", "sysErr")
	log.Println(logMsg)
}
