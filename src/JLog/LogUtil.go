package JLog

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logger *log.Logger
var loggerMap map[string]*log.Logger = make(map[string]*log.Logger)

func GetInstence(fileName string) *log.Logger {
	if loggerMap[fileName] != nil {
		return loggerMap[fileName]
	}
	if fileName != "" {
		logger = getFileAndConsoleLogger(fileName)
	} else {
		logger = getConsoleLogger()
	}
	loggerMap[fileName] = logger
	return logger
}

func getFileLogger(filename string) *log.Logger {
	dirString := "log/"
	filepath := "log/" + filename + ".log"
	out, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err.Error())
		err := os.MkdirAll(dirString, 0777)
		if err == nil {
			fmt.Println("create ok")
		}
	}
	return log.New(out, "[jLog]", log.Ldate|log.Ltime|log.Lshortfile)
}
func getConsoleLogger() *log.Logger {
	return log.New(os.Stdout, "[jLog]", log.Ldate|log.Ltime|log.Lshortfile)
}
func getFileAndConsoleLogger(filename string) *log.Logger {
	dirString := "log/"
	filepath := "log/" + filename + ".log"
	out, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err.Error())
		err := os.MkdirAll(dirString, 0777)
		if err == nil {
			fmt.Println("create ok")
		}
	}
	writers := io.MultiWriter(out, os.Stdout)
	return log.New(writers, "[jLog]", log.Ldate|log.Ltime|log.Lshortfile)

}
