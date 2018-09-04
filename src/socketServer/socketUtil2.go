package socketServer

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
	"yxtTool"
	"JLog"
)

var TASK_CHANNEL = make(chan string, 1024)

//客户端连接集合
var clientList = make(map[string]net.Conn, 1024)

var port = "6388"
var ch_change_so = make(chan map[string]net.Conn)

const MAX_COUNT int = 100

var SocketList map[string]interface{} = make(map[string]interface{})

func init() {
	port = yxtTool.ReadProperty()["port"]
}

func Listen() {

	ip := yxtTool.GetLoaclIp()

	fmt.Println("port" + port)
	addr, err := net.ResolveTCPAddr("tcp", ip+":"+port)
	if err != nil {
		JLog.PrintSysErr(err.Error())
		os.Exit(0)
	}

	server, err := net.ListenTCP("tcp", addr)
	if err != nil {
		JLog.PrintSysErr(err.Error())
		os.Exit(0)
	}
	defer server.Close()
	JLog.PrintSysErr("监听启动，监听地址：" + addr.IP.String() + ":" + strconv.Itoa(addr.Port))

	var cur_count_num = 0
	conn_chan := make(chan net.Conn)
	ch_change_chan := make(chan int)


	go func() {
		for conn_chan := range ch_change_chan {
			cur_count_num += conn_chan
			JLog.PrintSysErr("当前连接数：" + strconv.Itoa(cur_count_num))
		}
	}()

	for i := 0; i < MAX_COUNT; i++ {
		go func() {
			for conn := range conn_chan {
				ch_change_chan <- 1
				readMsg(conn)
				ch_change_chan <- 1
			}
		}()
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			continue
		}
		conn_chan <- conn
	}
}




/**
读数据
*/
func readMsg(conn net.Conn) (code string) {


	defer conn.Close()

	for {
		fmt.Println("wait to read............... ")
		b := make([]byte, 1024, 2048)
		err := conn.SetReadDeadline(time.Now().Add(time.Second * 20))
		count, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				JLog.PrintSysErr("连接断开:" + conn.RemoteAddr().String())
				return
			} else {
				JLog.PrintSysErr("读数据错误:" + err.Error())
				return
			}
		} else {
			b=b[0:count]
		}
	}
}

