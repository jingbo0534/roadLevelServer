package main

import (
	"net/http"
	"io"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"ServierBean"
	"JLog"
	"os"
	"static"
	"yxtTool"
	"strings"
	"time"
	"github.com/muesli/cache2go"
	"strconv"
)

var i=0
var sendChanel=make(chan ServierBean.SendBean,1024)
var key = yxtTool.ReadProperty()["mapKey"]
var cache = cache2go.Cache("myCache")
var local, _ = time.LoadLocation("Local")




func main() {
	defer delFile()
	out,_:= os.OpenFile(static.RoadLevelFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	out.Write([]byte("ok"))
	out.Close()
	http.HandleFunc("/", hello)
	http.HandleFunc("/alarmVoice", voice)
	http.HandleFunc("/driverLong",driverLongTime)
	http.HandleFunc("/fence",fence)
	http.HandleFunc("/speedZero",speedZero)

	http.ListenAndServe(":8080", nil)
}


func hello(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body) //此处可增加输入过滤
	if err != nil {
		fmt.Println("POST请求:读取body失败", err)
		return
	}
	s:=&ServierBean.SendBean{}
	err=json.Unmarshal(body,s)
	if err!=nil{
		fmt.Println(err)
	}
	JLog.WriteClientMsg(string(body),s.IDC,"log")
	go ServierBean.InTroubleRoad(s)
	if s.GPS_Speed<10{
		//fmt.Println("速度太低了。。。")
	}else{
		ServierBean.HasOverSpeed(s)
	}

	//根行为据速度进行驾驶判断
	//s.GetEventThing()






	io.WriteString(w, "Hello world!")
}







//紧急报警服务程序
func voice(w http.ResponseWriter, r *http.Request) {
	urlStr:=r.RequestURI
	urlStr=strings.Replace(urlStr,"/alarmVoice?","",1)
	strs:=strings.Split(urlStr,",")
	fmt.Println(urlStr)
	type SendBean struct {
		Cid string `json:"cid"`
		Poi string
		Time string
	}
	sendBean:=SendBean{}
	sendBean.Cid=strs[2]
	sendBean.Poi=strs[0]+","+strs[1]
	sendBean.Time=time.Now().Format(static.TimeLayOut1)
	b,_:=json.Marshal(sendBean)
	sendRs,_:=yxtTool.Http_Post(static.AlarmVoiceUrl,string(b))
	fmt.Println(sendRs)
}

/**
疲劳驾驶
 */
func driverLongTime(w http.ResponseWriter, r *http.Request)  {
	type sendBean struct {
		Obd string
		Cid string
		Stime string
		TimeLong string
		Lon string
		Lat string
		GpsSpeed float64
		PositionTime string
	}
	body, err := ioutil.ReadAll(r.Body) //此处可增加输入过滤
	if err != nil {
		fmt.Println("POST请求:读取body失败", err)
		return
	}
	s:=&sendBean{}
	err=json.Unmarshal(body,s)
	if err!=nil{
		fmt.Println(err)
	}
	//JLog.WriteClientMsg(string(body),s.obd,"log")
	//fmt.Println(string(body))
	if getOverTime(s.Obd)==""{
		fmt.Println("------------------------疲劳驾驶了，请靠边停车休息:"+s.Obd+" "+s.TimeLong)
		JLog.WriteClientMsg("\t"+s.Stime+"\t"+time.Now().Format(static.TimeLayOut1)+"疲劳驾驶",s.Obd,"timeLong")

		//下发语音报警提醒
		msg := yxtTool.MergeString("voiceobdno=",s.Obd,"&voicecid=",s.Cid,"&voicemsg=")
		msg = yxtTool.MergeString(msg,"您已连续驾驶：", s.TimeLong,"小时")
		if static.IDCIn(s.Obd) {
			sendRs:=yxtTool.DoGet(static.RoadVoiceUrl,msg)
			fmt.Println(msg+" "+sendRs)
			JLog.WriteClientMsg( "\t"+" 下发播报："+msg ,s.Obd,"timeLong")

			beanTemp:=&ServierBean.SendBean{}
			beanTemp.Cid=s.Cid
			lon,_:=strconv.ParseFloat(s.Lon,64)
			lat,_:=strconv.ParseFloat(s.Lat,64)
			beanTemp.Point=ServierBean.Point{lon,lat}
			beanTemp.Position_time=s.PositionTime
			beanTemp.GPS_Speed=s.GpsSpeed
			ServierBean.SaveAlarm(beanTemp, "疲劳驾驶", s.TimeLong)

		}
		setOverTime(s.Obd,"1")
	}
}

//空转报警
func speedZero(w http.ResponseWriter, r *http.Request)  {
	type sendBean struct {
		Obd string
		Cid string
		Stime string
		TimeLong string
		Lon string
		Lat string
		GpsSpeed float64
		PositionTime string
	}
	body, err := ioutil.ReadAll(r.Body) //此处可增加输入过滤
	if err != nil {
		fmt.Println("POST请求:读取body失败", err)
		return
	}
	s:=&sendBean{}
	err=json.Unmarshal(body,s)
	if err!=nil{
		fmt.Println(err)
	}
	if s.GpsSpeed < 2{
		stime:=getCache(s.Obd+"speedZero")
		if stime==""{
			setSpeedZero(s.Obd,time.Now().Format(static.TimeLayOut1))
		}else{
			timeGps, _ := time.ParseInLocation(static.TimeLayOut1, stime, local)
			poistionTime, _ := time.ParseInLocation(static.TimeLayOut1, s.PositionTime, local)


			//  当前点时间和上个点时间差值 小于 1分钟
			if time.Now().Sub(poistionTime) < time.Minute{

				//0速度持续超过3 分钟
				if time.Now().Sub(timeGps).Minutes() > time.Minute.Minutes()*10 {

					//语音播报
					msg := yxtTool.MergeString("voiceobdno=",s.Obd,"&voicecid=",s.Cid,"&voicemsg=")
					msg = yxtTool.MergeString(msg, "空转报警")
					
					if static.IDCIn(s.Obd)   {
						if getZeroSpeed(s.Obd)==""{ //播报间隔
							sendRs:=yxtTool.DoGet(static.RoadVoiceUrl,msg)
							fmt.Println(msg+" "+sendRs)
							JLog.WriteClientMsg( "\t"+" 下发播报："+msg ,s.Obd,"zeroSpeed")

							beanTemp:=&ServierBean.SendBean{}
							beanTemp.Cid=s.Cid
							lon,_:=strconv.ParseFloat(s.Lon,64)
							lat,_:=strconv.ParseFloat(s.Lat,64)
							beanTemp.Point=ServierBean.Point{lon,lat}
							beanTemp.Position_time=s.PositionTime
							beanTemp.GPS_Speed=s.GpsSpeed
							ServierBean.SaveAlarm(beanTemp, "空转报警", "")
							setZeroSpeed(s.Obd,"1")
						}else {
						}
					}
				}
			}else{
				setSpeedZero(s.Obd,time.Now().Format(static.TimeLayOut1))
			}
		}
	}else{
		//时间重置
		setSpeedZero(s.Obd,time.Now().Format(static.TimeLayOut1))
	}
}




/**
疲劳驾驶
 */
func fence(w http.ResponseWriter, r *http.Request)  {
	type sendBean struct {
		Cid string
		Obd string
		STime string
		AlarmType string
		FenceName string
		Lon string
		Lat string
	}
	body, err := ioutil.ReadAll(r.Body) //此处可增加输入过滤
	if err != nil {
		fmt.Println("POST请求:读取body失败", err)
		return
	}
	s:=&sendBean{}
	err=json.Unmarshal(body,s)
	if err!=nil{
		fmt.Println(err)
	}


	msg := yxtTool.MergeString("voiceobdno=",s.Obd,"&voicecid=",s.Cid,"&voicemsg=")
	msg = yxtTool.MergeString(msg, "您已驶入禁行区域")

	if static.IDCIn(s.Obd) {
		sendRs:=yxtTool.DoGet(static.RoadVoiceUrl,msg)
		fmt.Println(msg+" "+sendRs)
		JLog.WriteClientMsg( "\t"+" 下发播报："+msg ,s.Obd,"timeLong")
	}

	JLog.WriteClientMsg(string(body),s.Obd,"log")
	JLog.WriteClientMsg("\t"+s.STime+"\t"+s.AlarmType+"\t"+s.FenceName,s.Obd,"eleFence")
	fmt.Println("围栏报警:"+s.AlarmType+" "+s.FenceName+" "+s.STime)
}


func setSpeedZero(key string,value string) {
	cache.Add(key+"speedZero",0,value)
}
func getCache(key string) string {
	res, err :=cache.Value(key)
	if err!=nil{
		fmt.Println(err)
		return ""
	}
	return res.Data().(string)
}

func setZeroSpeed(key string,value string) {
	cache.Add(key+"ZeroSpeed",time.Minute*8,value)
}

func setOverTime(key string,value string) {
	cache.Add(key+"overTime",time.Minute*12,value)
}

func getOverTime(key string) string {
	return getWithOverTime(key+"overTime")
}
func getZeroSpeed(key string) string {
	return getWithOverTime(key+"ZeroSpeed")
}

func getWithOverTime(key string) string {
	res, err :=cache.Value(key)
	if err!=nil{
		fmt.Println(err)
		return ""
	}
	if  res.LifeSpan()!=0 && time.Now().Sub(res.CreatedOn())>res.LifeSpan(){
		fmt.Println("out of life time")
		cache.Delete(key)
		return  ""
	}
	return res.Data().(string)
}


func delFile()  {
	file, err := os.Open(static.RoadLevelFile)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	if file!=nil{
		err:=os.Remove(static.RoadLevelFile)
		if err!=nil{
			fmt.Println("删除文件失败")
		}else{
			fmt.Println("删除文件 OK")
		}
	}
	return
}