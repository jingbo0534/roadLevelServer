package main

import (

	"os"
	"fmt"
	"log"
	"bufio"

	"strings"
	"strconv"
)

var (
	textMap map[string]int
	addText chan string
)
func main()  {
	textMap=make(map[string]int)
	addText=make(chan string)


	path:="C:\\Users\\boger\\Desktop\\localhost_access_log.2018-05-10.txt"
	file,err:=os.Open(path)
	if err!=nil{
		log.Println(err)
	}
	if file !=nil {
		go func() {
			for str:=range addText{
				addKey:=strings.Split(str," ")[6]
				index:=strings.Index(addKey,"?")
				if index >0{
					addKey=addKey[0:index]
				}
				isIn:=false
				for key,_:=range textMap{
					if addKey==key{
						isIn=true;
					}
				}
				if isIn{
					value:=textMap[addKey]
					value++
					textMap[addKey]=value
				}else {
					textMap[addKey]=1
				}
			}
		}()


		scan:=bufio.NewScanner(file)
		for scan.Scan(){
			r_txt:=scan.Text()
			addText <- r_txt
			//log.Println(r_txt)
		}
		for key,value:=range textMap{
			fmt.Println(key+"\t"+strconv.Itoa(value))
		}


	}
}

