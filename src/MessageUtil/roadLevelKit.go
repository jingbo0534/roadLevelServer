package MessageUtil

import (
	"container/list"
	"encoding/json"
	"fmt"
	"static"
	"strconv"
	"time"
	"yxtTool"
)

var quest_url = "http://restapi.amap.com/v3/autograsp"
var key = "948ce662e5e4d0707dd1797a3736951e"

func init() {
	quest_url = yxtTool.ReadProperty()["roadLevelURI"]
	key = yxtTool.ReadProperty()["mapKey"]
}

type RoadLevel struct {
	Status,
	Count,
	Info,
	InfoCode string
	Roads []*RoadsItem
}

type RoadsItem struct {
	RoadName,
	CrossPoint,
	RoadLevel,
	MaxSpeed,
	IntersectionDistance string
	Intersection []interface{}
}

/**
test
*/
func Road() {
	s := []string{"108.837275,37.634455", "2018-04-03 09:59:56", "262", "55", "201236445"}
	m := make(map[string]string)
	m["onumber"] = s[4]
	m["OP"] = s[0]
	m["otime"] = s[1]
	m["direction"] = s[2]
	m["gps_speed"] = s[3]








	s1 := []string{"108.837445,37.634468", "2018-04-03 09:59:55", "260", "54"}
	s2 := []string{"108.837618,37.634491", "2018-04-03 09:59:54", "262", "54"}
	s3 := []string{"108.837788,37.634507", "2018-04-03 09:59:53", "260", "53"}
	s4 := []string{"108.837954,37.634527", "2018-04-03 09:59:52", "260", "51"}
	s5 := []string{"108.838114,37.634537", "2018-04-03 09:59:51", "158", "49"}
	s6 := []string{"108.838267,37.634561", "2018-04-03 09:59:50", "256", "48"}
	s7 := []string{"108.838418,37.634589", "2018-04-03 09:59:49", "256", "47"}
	s8 := []string{"108.838556,37.634608", "2018-04-03 09:59:48", "254", "47"}
	m1 := toMap(s1)
	m2 := toMap(s2)
	m3 := toMap(s3)
	m4 := toMap(s4)
	m5 := toMap(s5)
	m6 := toMap(s6)
	m7 := toMap(s7)
	m8 := toMap(s8)
	l := list.New()
	l.PushBack(m8)
	l.PushBack(m7)
	l.PushBack(m6)
	l.PushBack(m5)
	l.PushBack(m4)
	l.PushBack(m3)
	l.PushBack(m2)
	l.PushBack(m1)
	roadItems := GetRoadSpeed(l, m)
	for _, item := range roadItems {
		fmt.Println(*item)
	}
}

func toMap(s []string) map[string]string {
	m := make(map[string]string)
	m["OP"] = s[0]
	m["otime"] = s[1]
	m["direction"] = s[2]
	m["gps_speed"] = s[3]
	return m
}

/**
 *
 * @param qds
 *            坐标点数组对象，按时间升序排列
 * @param hds
 *            当前点数组
 * @return
 * @throws Exception
 */
func GetRoadSpeed(qds *list.List, hds map[string]string) []*RoadsItem {

	if quest_url == "" {
		return nil
	}
	params := getRestParameter(qds, hds)
	params = yxtTool.MergeString("key=", key, params)
	jsonStr := yxtTool.DoGet(quest_url, params)

	//jsonStr = strings.Replace(jsonStr, "[]", "\"\"", -1)
	//fmt.Println(jsonStr)
	//JLog.PrintSysErr(params);
	rl := &RoadLevel{}
	err := json.Unmarshal([]byte(jsonStr), rl)
	if err != nil {
		//fmt.Print("roldLevelError(L98):\t")
		fmt.Println(err.Error())
	}
	return rl.Roads
}

func getRestParameter(qds *list.List, hds map[string]string) string {
	parameter := yxtTool.MergeString("&output=JSON&carid=c3e9")
	parameter = yxtTool.MergeString(parameter, hds["onumber"]) // carid=key后四位+car唯一标识，这里填写OBD编号
	location := "&locations="
	timestamp := "&time="
	direction := "&direction="
	speed := "&speed="

	index := 0
	for e := qds.Front(); e != nil; e = e.Next() {
		//if index%5 == 0 {
			item := e.Value.(map[string]string)
			location = yxtTool.MergeString(location, item["OP"], "|")                  // 坐标，用','隔开，前面经度，后面维度
			timestamp = yxtTool.MergeString(timestamp, getUTCTime(item["otime"]), ",") //时间
			direction = yxtTool.MergeString(direction, item["direction"], ",")         //角度
			speed = yxtTool.MergeString(speed, item["gps_speed"], ",")                 //车速
		//}
		index++
	}
	location = yxtTool.MergeString(location, hds["OP"])                  // 坐标，用','隔开，前面经度，后面维度
	timestamp = yxtTool.MergeString(timestamp, getUTCTime(hds["otime"])) //时间
	direction = yxtTool.MergeString(direction, hds["direction"])         //角度
	speed = yxtTool.MergeString(speed, hds["gps_speed"])                 //车速

	parameter = yxtTool.MergeString(parameter, location)
	parameter = yxtTool.MergeString(parameter, timestamp)
	parameter = yxtTool.MergeString(parameter, direction)
	parameter = yxtTool.MergeString(parameter, speed)

	return parameter
}

func getUTCTime(time8 string) string {

	t, _ := time.Parse(static.TimeLayOut1, time8)
	utcTime := t.Unix()
	timestamp := strconv.FormatInt(utcTime, 10)
	return timestamp
}
