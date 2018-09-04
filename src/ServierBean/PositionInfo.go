package ServierBean

import (
	"yxtTool"
	"time"
	"static"
	"sqlUtil"
	"strconv"
	"math"
	"fmt"
)

type SendBean struct {
	GPS_Speed       float64
	Orientation     int
	Position_time   string
	Point           Point
	Cid 			string
	TbName			string
	IDC				string
}

//判断急加速急减速 性能问题 未启用
func (info *SendBean) GetEventThing()  {

	timeNow,_:=time.ParseInLocation(static.TimeLayOut1,info.Position_time,local)

	baseSql:=yxtTool.MergeString("select o.gps_speed,o.direction from ",obdDataBase,".",info.TbName," o where o.otime= ? ")

	if !info.getEventThing(timeNow,baseSql,1){
		if !info.getEventThing(timeNow,baseSql,2){
			info.getEventThing(timeNow,baseSql,3)
		}
	}
	if !info.getEventThing(timeNow,baseSql,-1){
		if !info.getEventThing(timeNow,baseSql,-2){
			info.getEventThing(timeNow,baseSql,-3)
		}
	}
}



func (info *SendBean) getEventThing(timeNow time.Time,baseSql string,timeDiffer int ) bool {
	nextPoint:= getOtherPoint(timeNow,timeDiffer,baseSql)
	if nextPoint!=nil {
		provSpeed:=yxtTool.StringToFloat(nextPoint["gps_speed"])
		//如果 时间差 >0 表示 上个点 - 当前点
		//如果 时间差 <0 表示 当前点 - 下一个
		speedDiffer:=0.0
		if timeDiffer > 0{
			speedDiffer=provSpeed-info.GPS_Speed
		}
		if timeDiffer > 0{
			speedDiffer=info.GPS_Speed-provSpeed
		}


		if isSpeedUp(speedDiffer,math.Abs(float64(timeDiffer)))==1{
			//急加速
			SaveEventInfo("急加速",info)
			return true
		}

		if isSpeedUp(speedDiffer,math.Abs(float64(timeDiffer)))==-1{
			//急减速
			SaveEventInfo("急减速",info)
			return true
		}
	}
	return false
}

//获得当前秒的 前后指定秒的 数据
func getOtherPoint(timeNow time.Time,addtimes int,sql string)map[string]string  {
	provTime:=timeNow.Add(time.Second*time.Duration(addtimes)).Format(static.TimeLayOut1)
	provList,_:=sqlUtil.QueryList(sql,provTime)
	if provList.Len() > 0 {
		provPoint:=provList.Front().Value.(map[string]string)
		return provPoint
	}else{
		return nil
	}
}
/**
  返回值 1急加速 -1 急减速 0 未触发
	 thisPoint  和 和provPoint  连个点按照时间顺序

 */
//判断急加速
func isSpeedUp(speedDiffer float64,timeDiffer float64)  int   {

	if speedDiffer >= (yxtTool.StringToFloat(speedUp) * timeDiffer){
		fmt.Println(yxtTool.FormatFloat(speedDiffer,1))
		return 1
	}
	if speedDiffer <= (yxtTool.StringToFloat(speedDown)*timeDiffer*-1) {
		fmt.Println(yxtTool.FormatFloat(speedDiffer,1))
		return -1
	}
	return 0
}


/**
存储报警信息
*/
func SaveEventInfo( eType string,info *SendBean) {


	if info.Point.Longitude<1 || info.Point.Latitude<1{
		return
	}

	lon := strconv.FormatFloat(info.Point.Longitude, 'f', -1, 64)
	lat := strconv.FormatFloat(info.Point.Latitude, 'f', -1, 64)
	position := yxtTool.MergeString(lon, ",", lat)



	gpsSpeed := strconv.FormatFloat(info.GPS_Speed, 'f', -1, 64)
	insertSql := yxtTool.MergeString("INSERT INTO ", dataBase, ".veh_event (`event_name`,`event_car`,`event_desc`,`event_position`,`event_speed`,`event_time`,`event_length`,`event_flag`) VALUES(? , ? , ? , ? , ? , ? , ? , ? )")
	sqlUtil.Insert(insertSql, eType, info.Cid, "1", position, gpsSpeed, info.Position_time, 0, 1)
}
