package yxtTool

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"net/url"
)

func DoGet(u string, params string) string {
	sendUrl := MergeString(u, "?", params)

	parseUrl,_:=url.Parse(sendUrl)
	parseUrl.RawQuery=parseUrl.Query().Encode()
	sendUrl=parseUrl.String()
	resp, err := http.Get(sendUrl)

	if err != nil {
		fmt.Print("HttpToolKit.go(13)")
		fmt.Println(err.Error())
		return ""
	}
	reader := resp.Body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Print("HttpToolKit.go(20)")
		fmt.Println(err.Error())
		return ""
	}
	body := string(b)
	//fmt.Print("HttpToolKit.go(26)")
	//fmt.Println(body)
	return body

}


func Http_Post(url string, bodyStr string)(rs string , err error) {
	//post的body内容,当前为json格式
	reqbody :=bodyStr
	rs=""
	//创建请求
	postReq, err := http.NewRequest("POST",
		url, //post链接内容
		strings.NewReader(reqbody)) //post
	if err != nil {
		fmt.Println("POST请求1:创建请求失败", err)
		return
	}
	//增加header
	postReq.Header.Set("Content-Type", "application/json; encoding=utf-8")

	//执行请求
	client := &http.Client{}
	resp, err := client.Do(postReq)
	if err != nil {
		fmt.Println("POST请求2:创建请求失败", err)
		return rs ,err
	} else {
		//读取响应
		body, err := ioutil.ReadAll(resp.Body) //此处可增加输入过滤
		if err != nil {
			fmt.Println("POST请求3:读取body失败", err)
			return rs,err
		}
		rs=string(body)
		return rs,nil
	}
	defer resp.Body.Close()

	return
}