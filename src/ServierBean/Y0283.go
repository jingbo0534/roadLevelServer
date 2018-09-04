package ServierBean

import (
	"JLog"
	"MessageUtil"
	"container/list"
	"fmt"
	"github.com/muesli/cache2go"
	"sqlUtil"
	"static"
	"strconv"
	"strings"
	"time"
	"yxtTool"
)

/**
GPS 指令上传
*/
var (
	dataBase     string
	obdDataBase  string
	roadLevelURI string
	sendChanel   = make(chan SendBean)
	cache        = cache2go.Cache("myCache")
	local, _     = time.LoadLocation("Local")
	speedUp      string
	speedDown    string
)

func init() {
	dataBase = yxtTool.ReadProperty()["dataBase"]
	obdDataBase = yxtTool.ReadProperty()["obdDataBase"]
	roadLevelURI = yxtTool.ReadProperty()["roadLevelURI"]
	speedUp = yxtTool.ReadProperty()["speedUp"]
	speedDown = yxtTool.ReadProperty()["speedDown"]
}

func cacheSpeedTime(key string) {
	cache.Add(key+"overSpeedTime", 0, time.Now().Format(static.TimeLayOut1))
}
func getSpeedTime(key string) string {
	res, err := cache.Value(key)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return res.Data().(string)
}

func cacheEleName(key string, value string) {
	cache.Add(key+"eleName", 0, value)
}
func getEleName(key string) string {
	res, err := cache.Value(key + "eleName")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return res.Data().(string)
}

/**
报警设置
*/
func setCache(key string, value string) {
	cache.Add(key, time.Second*20, value)
}

func getCache(key string) string {
	res, err := cache.Value(key)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	if res.LifeSpan() != 0 && time.Now().Sub(res.CreatedOn()) > res.LifeSpan() {
		fmt.Println("out of life time")
		cache.Delete(key)
		return ""
	}
	return res.Data().(string)
}

func cacheAlarmTime(key string, stime string) {
	cache.Add(key+"alarmTime", 0, stime)
}

func getAlarmTimeCache(key string) string {
	res, err := cache.Value(key + "alarmTime")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return res.Data().(string)
}

func cacheEles(key string, eles list.List) {
	cache.Add(key+"eles", 0, eles)
}
func getEles(key string) list.List {
	res, err := cache.Value(key + "eles")
	if err != nil {
		fmt.Println(err)
		return list.List{}
	}
	return res.Data().(list.List)
}


func HasOverSpeed(info *SendBean){


	if !RoadLevelSpeed(info){ //道路等级抓路失败，则采用电子围栏
		EleFence2(info)
	}
}


/*
入参 : carId
当前速度
当前时间
当前坐标
方向
*/

/**
道逻路等级报警辑函数
*/
func RoadLevelSpeed(info *SendBean) bool {
	timeGps, _ := time.ParseInLocation(static.TimeLayOut1, info.Position_time, local)
	if  info.GPS_Speed > 120 {
		//超速了
		SaveOverSpeedAlarm(info, "自定义超速报警", "", yxtTool.FloatToString(120, 0), "")
		//TODO 下发超速度报警 进行时间判断

		if time.Now().Sub(timeGps) > time.Second*10 {
			JLog.WriteClientMsg("\t 120 \t播报延时超过 10 秒", info.IDC, "overSpeed")
			return true
		} else {
			info.TbName = yxtTool.FloatToString(120, 0)
			SendOverSpeedVoice(*info)
			return true
		}
	}

	//查询定的车辆绑规则
	roadLevel_sql := yxtTool.MergeString("select rl.* from ")
	roadLevel_sql = yxtTool.MergeString(roadLevel_sql, dataBase, ".vehicle_basic vb left join ")
	roadLevel_sql = yxtTool.MergeString(roadLevel_sql, dataBase, ".roadlevel rl on vb.`vehicle_roadlevel`=rl.`rl_id` where vb.`vehicle_id`=?")
	roadLevel_sql = yxtTool.MergeString(roadLevel_sql, " and vb.`vehicle_flag`=1 and  rl.`rl_id` is not null")
	roadRule, _ := sqlUtil.QueryList(roadLevel_sql, info.Cid)

	if roadRule != nil && roadRule.Len() > 0 {
		rule := roadRule.Front().Value.(map[string]string)

		//查询最近的40个点
		sel5Point := yxtTool.MergeString("select CONCAT (obd.longitude , ',' ,obd.latitude) AS OP ,obd.otime,obd.direction,obd.gps_speed from ")
		sel5Point = yxtTool.MergeString(sel5Point, obdDataBase, ".", info.TbName, " obd WHERE obd.`otime`< ? ORDER BY otime DESC LIMIT 0,10" )
		ps, _ := sqlUtil.QueryList(sel5Point, info.Position_time)

		//  判断当前点跟上个点时间差

		prev:= ps.Front().Value.(map[string]string)
		prevPositionTime:=prev["otime"]
		prevTime,_:=time.ParseInLocation(static.TimeLayOut1,prevPositionTime,local)
		curPoiTime,_:=time.ParseInLocation(static.TimeLayOut1,info.Position_time,local)
		if curPoiTime.Sub(prevTime).Seconds() > 5{
			//当前点跟上个点差值大于 5S 不进行抓路
			return false
		}

		//集合翻转  点按递增时间
		ls := list.New()

		for e := ps.Front(); e != nil; e = e.Next() {
			ls.PushFront(e.Value)
		}

		//根据 查询到的四个点 查找该车所在的道路的等级
		var hds = make(map[string]string)
		lon := strconv.FormatFloat(info.Point.Longitude, 'f', -1, 64)
		lat := strconv.FormatFloat(info.Point.Latitude, 'f', -1, 64)
		hds["onumber"] = info.IDC
		hds["OP"] = yxtTool.MergeString(lon, ",", lat)
		hds["otime"] = info.Position_time
		hds["direction"] = strconv.Itoa(info.Orientation)
		hds["gps_speed"] = strconv.FormatFloat(info.GPS_Speed, 'f', -1, 64)
		roadItems := MessageUtil.GetRoadSpeed(ls, hds)
		roadSpeed := "-1"

		if roadItems != nil && len(roadItems) > 0 {
			roadItem := roadItems[0]
			for i := 0; i < len(roadItems); i++ {
				if roadItems[i].RoadName != "" {
					roadItem=roadItems[i]
					for j := 0; j < len(roadItems); j++ {
						if yxtTool.StringToFloat(roadItems[j].MaxSpeed)  >= yxtTool.StringToFloat(roadItem.MaxSpeed) {
							// 判断去限速最高的一条路，如果所有限速都一样，则取最后一条数据（根据最后一个点抓到的路点）
							roadItem=roadItems[j]
						}
					}
					break
				}
			}
			if roadItem.RoadName == "" {
				//if static.IDCIn(info.IDC) {
				JLog.WriteClientMsg("\t"+info.Point.Lon()+","+info.Point.Lat()+"", "road", "overSpeed"+info.IDC)
				//}
				return false
			}

			roadSpeed=roadItem.MaxSpeed+","+strconv.Itoa(int(yxtTool.StringToFloat(roadItem.MaxSpeed)*0.8))
			if  strings.Contains(roadItem.RoadName,"高速"){
				roadSpeed="120,96"
			}
			if roadSpeed != "-1" {

				st := rule["rl_start_time"]
				et := rule["rl_end_time"]

				selRoad := "select * from " + dataBase + ".specialroad s where s.sr_name=?"
				specilRoads, _ := sqlUtil.QueryList(selRoad, roadItem.RoadName)
				if specilRoads.Len() > 0 {
					sr := specilRoads.Front().Value.(map[string]string)
					roadSpeed = sr["sr_speed"]
					st = sr["sr_stime"]
					et = sr["sr_dtime"]
				}else{
					//没有设定道路规则
					return false
				}


				no := timeGps.Format(static.TimeLayOut4)

				isNight := timeIn(st, et, no)
				if isNight {
					//判断在夜间时段 执行夜间时段限速规则
					if len(strings.Split(roadSpeed, ",")) < 2 {
						fmt.Println("没有夜间限速")
						return false
					}
					roadSpeed = strings.Split(roadSpeed, ",")[1]
				} else {
					roadSpeed = strings.Split(roadSpeed, ",")[0]
				}
				maxSpeed, err := strconv.ParseFloat(roadSpeed, 64)

				roadinfo := info.IDC + " " + roadItem.RoadName + " " + roadItem.RoadLevel + " " + roadItem.MaxSpeed + " " + yxtTool.FloatToString(maxSpeed, 0) + " " + yxtTool.FloatToString(info.GPS_Speed, 0) + " " + time.Now().Sub(timeGps).String()
				//if static.IDCIn(info.IDC) {
					fmt.Println(roadinfo)
				//}
				JLog.WriteClientMsg("\t"+roadinfo+"", info.IDC, "overSpeed")
				if err == nil && info.GPS_Speed > maxSpeed {
					//超速了
					SaveOverSpeedAlarm(info, "自定义超速报警", roadItem.RoadName, yxtTool.FloatToString(maxSpeed, 0), roadItem.RoadLevel)
					//TODO 下发超速度报警 进行时间判断

					if time.Now().Sub(timeGps) > time.Second*10 {
						JLog.WriteClientMsg("\t"+roadinfo+"\t播报延时超过 10 秒", info.IDC, "overSpeed")
						return true
					} else {
						info.TbName = yxtTool.FloatToString(maxSpeed, 0)
						SendOverSpeedVoice(*info)
						return true
					}
				} else {
					//没超速
					resetTimeLong(*info)
					return true
				}
			}
		}
	}
	return false
}

/**
围栏逻辑判断

做道路超速用，只判断围栏超速报警

*/
func EleFence2(info *SendBean) bool {
	//查询电子围栏信息
	//易公务新数据库结构
	ele_fence_sql := yxtTool.MergeString("SELECT ele_fence.* FROM ")
	ele_fence_sql = yxtTool.MergeString(ele_fence_sql, dataBase, ".vehicle_basic inner JOIN ")
	ele_fence_sql = yxtTool.MergeString(ele_fence_sql, dataBase, ".ef_car ef_car ON vehicle_basic.`vehicle_id` = ef_car.`ef_car_cid`  inner JOIN ")
	ele_fence_sql = yxtTool.MergeString(ele_fence_sql, dataBase, ".ele_fence  ON ef_car.`ef_car_eid` = ele_fence.`ef_id` ")
	ele_fence_sql = yxtTool.MergeString(ele_fence_sql, " WHERE ele_fence.`ef_trigger`= '道路限速' AND vehicle_basic.vehicle_id=?  ORDER BY ele_fence.ef_maxspeed DESC")

	var fenceList list.List
	fenceList = getEles(info.IDC)

	if fenceList.Len() < 1 {
		eles, _ := sqlUtil.QueryList(ele_fence_sql, info.Cid)
		fenceList = *eles
		cacheEles(info.IDC, fenceList)
	} else {
		fenceList = getEles(info.IDC)
	}

	//循环判断围栏
	flag := false
	for e := fenceList.Front(); e != nil; e = e.Next() {
		fence := e.Value.(map[string]string)
		flag = info.Point.PtInEleFence(fence)
		st := fence["ef_starttime"]
		et := fence["ef_stoptime"]
		ruleSpeedStr := fence["ef_maxspeed"]

		//点在围栏内的情况
		if flag {
			//进区域
			eleName := getEleName(info.IDC)
			fmt.Println( info.IDC+"\t" + eleName + "\t" + fence["ef_name"])
			if eleName == "" || eleName != fence["ef_name"] {
				msg := yxtTool.MergeString("voiceobdno=", info.IDC, "&voicecid=", info.IDC, "&voicemsg=")
				msg = yxtTool.MergeString(msg, "您已进入限速区域，限速：", fence["ef_maxspeed"])
				//yxtTool.DoGet(static.RoadVoiceUrl,msg)
				fmt.Println(msg)
				cacheEleName(info.IDC, fence["ef_name"])
			}

			overSpeeded := false

			timeGps, _ := time.ParseInLocation(static.TimeLayOut1, info.Position_time, local)
			no := timeGps.Format(static.TimeLayOut4)

			isNight := timeIn(st, et, no)

			if ruleSpeedStr == "" {
				return flag
			}
			ruleSpeed := yxtTool.StringToFloat(ruleSpeedStr)
			realSpeed:=ruleSpeed

			if isNight {
				// 夜间限速
				realSpeed=ruleSpeed*0.8

				if info.GPS_Speed > realSpeed {
					//限时 && 限速 &&超速
					overSpeeded = true
				} else {
					overSpeeded = false
				}
			} else {
				realSpeed = ruleSpeed
				if info.GPS_Speed > realSpeed {
					overSpeeded = true
				} else {
					overSpeeded = false
				}
			}

			if overSpeeded {

				local, _ := time.LoadLocation("Local")
				timeGps, _ := time.ParseInLocation(static.TimeLayOut1, info.Position_time, local)

				if time.Now().Sub(timeGps) > time.Second*10 {
					JLog.WriteClientMsg("\t播报延时超过 10 秒", info.IDC, "overSpeed")

				} else {
					info.TbName =yxtTool.FloatToString(realSpeed,0)
					SendOverSpeedVoice(*info)
				}
				SaveOverSpeedAlarm(info, "自定义超速报警", fence["ef_name"], yxtTool.FloatToString(realSpeed,0), "")
			}else {
				//没超速逻辑
				//重置长时间超速记录
				resetTimeLong(*info)
			}

			return flag
		}

	}
	return flag

}

/**
存储道路等级报警信息
*/
func SaveOverSpeedAlarm(info *SendBean, trigger string, content string, roodSpeed string, roadlevelnm string) {
	if info.Point.Latitude == 0 {
		return
	}
	lon := strconv.FormatFloat(info.Point.Longitude, 'f', -1, 64)
	lat := strconv.FormatFloat(info.Point.Latitude, 'f', -1, 64)
	position := yxtTool.MergeString(lon, ",", lat)
	gpsSpeed := strconv.FormatFloat(info.GPS_Speed, 'f', -1, 64)
	insertSql := yxtTool.MergeString("INSERT INTO ", dataBase, ".obd_alarm (onumber,alarm_time,alarm_type,alarm_content,alarm_position,gps_speed,veh_id,roadlevelnm,roadspeed) VALUES(? , ? , ? , ? , ? , ? , ? ,? , ? )")
	sqlUtil.Insert(insertSql, info.Cid, info.Position_time, trigger, content, position, gpsSpeed, info.Cid, roadlevelnm, roodSpeed)
}

/**
存储报警信息
*/
func SaveAlarm(info *SendBean, trigger string, content string) {
	if info.Point.Latitude == 0 {
		return
	}
	lon := strconv.FormatFloat(info.Point.Longitude, 'f', -1, 64)
	lat := strconv.FormatFloat(info.Point.Latitude, 'f', -1, 64)
	position := yxtTool.MergeString(lon, ",", lat)
	gpsSpeed := strconv.FormatFloat(info.GPS_Speed, 'f', -1, 64)
	insertSql := yxtTool.MergeString("INSERT INTO ", dataBase, ".obd_alarm (onumber,alarm_time,alarm_type,alarm_content,alarm_position,gps_speed,veh_id) VALUES(? , ? , ? , ? , ? , ? , ? )")
	sqlUtil.Insert(insertSql, info.Cid, info.Position_time, trigger, content, position, gpsSpeed, info.Cid)
}

/**
下发道路等级超速报警
*/
func SendOverSpeedVoice(info SendBean) {

	alarmTime, _ := time.Parse(static.TimeLayOut1, info.Position_time)
	alarmTime = alarmTime.Add(time.Second * 20 * -1)

	//记录连续报警
	hasLong := overSpeedLong(info)
	//报警间隔 10S
	if hasLong && getCache(info.IDC) == "" {
		//todo  下发语音报警提醒
		msg := yxtTool.MergeString("voiceobdno=", info.IDC, "&voicecid=", info.Cid, "&voicemsg=")
		msg = yxtTool.MergeString(msg, "您已超速，限速", info.TbName, "请安全驾驶")
		if static.IDCIn(info.IDC) {
			sendRs := yxtTool.DoGet(static.RoadVoiceUrl, msg)
			fmt.Println(msg + " " + sendRs)
		}
		JLog.WriteClientMsg("\t"+" 下发播报："+msg, info.IDC, "overSpeed")
		setCache(info.IDC, time.Now().Format(static.TimeLayOut1))
	}
}

func overSpeedLong(info SendBean) bool {

	speedTime := getSpeedTime(info.IDC)
	if speedTime == "" {
		cacheAlarmTime(info.IDC, time.Now().Format(static.TimeLayOut1))
		cacheSpeedTime(info.IDC)
		return true
	}
	stime := getAlarmTimeCache(info.IDC)
	if stime == "" {
		cacheAlarmTime(info.IDC, time.Now().Format(static.TimeLayOut1))
	} else {
		timeGps, _ := time.ParseInLocation(static.TimeLayOut1, stime, local)
		poistionTime, _ := time.ParseInLocation(static.TimeLayOut1, info.Position_time, local)
		provSpeedTime, _ := time.ParseInLocation(static.TimeLayOut1, speedTime, local)

		//  当前点时间和上个点时间差值 小于5秒
		if provSpeedTime.Sub(poistionTime).Seconds() < time.Second.Seconds()*5 {

			//0速度持续超过3 分钟
			if time.Now().Sub(timeGps).Seconds() > time.Second.Seconds()*30 {
				//严重超速
				alarmTime := time.Now().Format(static.TimeLayOut4)
				msg := yxtTool.MergeString("voiceobdno=", info.IDC, "&voicecid=", info.Cid, "&voicemsg=")
				msg = yxtTool.MergeString(msg, alarmTime, "您已长时超速", "当前速度"+yxtTool.FloatToString(info.GPS_Speed, 1))
				if static.IDCIn(info.IDC) {
					sendRs := yxtTool.DoGet(static.RoadVoiceUrl, msg)
					fmt.Println(msg + " " + sendRs)
					JLog.WriteClientMsg("\t"+" 下发播报："+msg, info.IDC, "overSpeed")
				}
				cacheAlarmTime(info.IDC, time.Now().Format(static.TimeLayOut1))
				setCache(info.IDC, time.Now().Format(static.TimeLayOut1))
				cacheSpeedTime(info.IDC)
				return false
			}
		} else {
			//时间重置
			cacheSpeedTime(info.IDC)
			cacheAlarmTime(info.IDC, time.Now().Format(static.TimeLayOut1))
		}
	}
	return true
}

func resetTimeLong(info SendBean)  {
	cacheSpeedTime(info.IDC)
	cacheAlarmTime(info.IDC, time.Now().Format(static.TimeLayOut1))
}


func timeIn(st string, et string, no string) bool {

	if st=="" || et=="" || no==""{
		return false
	}

	st = strings.Replace(st, ":", ".", 1)
	et = strings.Replace(et, ":", ".", 1)
	no = strings.Replace(no, ":", ".", 1)

	stTime := yxtTool.StringToFloat(st)
	etTime := yxtTool.StringToFloat(et)
	noTime := yxtTool.StringToFloat(no)

	isNight := false

	if stTime < etTime { //开始时间小于结束时间
		if stTime < noTime && noTime < etTime {
			//在晚间时段
			isNight = true
		} else {
			isNight = false
		}
	} else if stTime > etTime { //开始时间 大于结束时间，说明 跨天了
		if (stTime < noTime && noTime < 24.00) || (00.0 < noTime && noTime < etTime) {
			//在晚间时段
			isNight = true
		} else {
			isNight = false
		}
	}
	return isNight
}

func inFenceFlag(fence map[string]string, info *SendBean) (alarmStatus string, alarmType string) {
	st := fence["ef_starttime"]
	et := fence["ef_stoptime"]
	ruleSpeedStr := fence["ef_maxspeed"]
	if st != "" && et != "" { //有时段限制
		noTime, _ := time.Parse(static.TimeLayOut1, info.Position_time)
		no := noTime.Format(static.TimeLayOut4)
		isIn := timeIn(st, et, no)
		if isIn {
			//在限时范围内
			if ruleSpeedStr != "" {
				//限时 && 限速
				ruleSpeed := yxtTool.StringToFloat(ruleSpeedStr)
				if info.GPS_Speed > ruleSpeed {
					//限时 && 限速 &&超速
					//todo 1001
					alarmType = "围栏报警-限时限速"
					alarmStatus = "1001"
					return
				} else {
					alarmStatus = "1006"
					return
				}
			} else {
				//限时 不限速 在限时时段内 进入区域
				alarmType = "围栏报警-限时"
				alarmStatus = "1002"
				return
			}
		} else {
			alarmStatus = "1006"
			return
		}
	} else { // 不限时段，判断是否限速
		if ruleSpeedStr != "" {
			// 限速
			ruleSpeed := yxtTool.StringToFloat(ruleSpeedStr)
			if info.GPS_Speed > ruleSpeed {
				// 限速 && 超速
				//todo 报警 1003
				alarmType = "围栏报警-限速"
				alarmStatus = "1003"
				return
			} else {
				alarmStatus = "1006"
				return
			}
		} else {
			//不限时，不限速，进入区域报警
			//todo 报警 1004
			alarmType = "围栏报警-限域(进)"
			alarmStatus = "1004"
			return
		}
	}
}
