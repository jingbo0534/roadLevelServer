package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
)

func main() {
	http.HandleFunc("/", action1)
}

func action1(w http.ResponseWriter,r *http.Request)  {
	body, err := ioutil.ReadAll(r.Body) //此处可增加输入过滤
	if err != nil {
		fmt.Println("POST请求:读取body失败", err)
		return
	}
	fmt.Println(string(body))
}