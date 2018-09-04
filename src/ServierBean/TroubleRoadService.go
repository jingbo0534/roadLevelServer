package ServierBean

import (
	"time"
	"static"
	"JLog"
	"yxtTool"
	"sqlUtil"
	"fmt"
	"container/list"
)
func cacheToubleEles( eles list.List) {
	cache.Add("toubleEles", time.Hour, eles)
}
func getToubleEles(key string) list.List {
	res, err := cache.Value("toubleEles")
	if err != nil {
		fmt.Println(err)
		return list.List{}
	}
	return res.Data().(list.List)
}

func cacheInToubleEles(key string,ef_id string) {
	cache.Add(key+"intoubleEles", time.Hour, ef_id)
}
func getInToubleEles(key string) string {
	res, err := cache.Value(key+"intoubleEles")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return res.Data().(string)
}


/**
围栏逻辑判断

做道路超速用，只判断围栏超速报警

*/
func InTroubleRoad(info *SendBean) bool {
	//查询电子围栏信息
	//易公务新数据库结构
	ele_fence_sql := yxtTool.MergeString("SELECT * FROM ")
	ele_fence_sql = yxtTool.MergeString(ele_fence_sql, dataBase, ".ele_fence ")
	ele_fence_sql = yxtTool.MergeString(ele_fence_sql, " WHERE ele_fence.`ef_trigger`= '道路提醒' ")

	timeGps, _ := time.ParseInLocation(static.TimeLayOut1, info.Position_time, local)



	if time.Now().Sub(timeGps) > time.Second*10 {
		JLog.WriteClientMsg("\t播报延时超过 10 秒", info.IDC, "toubleRoadVoice")
		return false
	}



	var fenceList list.List
	fenceList = getToubleEles(info.IDC)

	if fenceList.Len() < 1 {
		eles, _ := sqlUtil.QueryList(ele_fence_sql, info.Cid)
		fenceList = *eles
		cacheToubleEles(fenceList)
	} else {
		fenceList = getEles(info.IDC)
	}

	//循环判断围栏
	flag := false
	for e := fenceList.Front(); e != nil; e = e.Next() {

		fence := e.Value.(map[string]string)
		flag = info.Point.PtInEleFence(fence)
		ef_id := fence["ef_id"]
		ef_remark := fence["ef_remark"]
		ef_name := fence["ef_name"]


		//点在围栏内的情况
		if flag {
			if getInToubleEles(info.IDC)!="" {
				fmt.Println( info.IDC+"\t 已经在：" + getInToubleEles(info.IDC))
				return false
			}
			cacheInToubleEles(info.IDC,ef_id)
			//区域内
			fmt.Println( info.IDC+"\t" + ef_name + "\t" + fence["ef_name"])
			msg := yxtTool.MergeString("voiceobdno=", info.IDC, "&voicecid=", info.IDC, "&voicemsg=")
			msg = yxtTool.MergeString(msg,ef_remark)
			yxtTool.DoGet(static.RoadVoiceUrl,msg)
			fmt.Println(msg)
			cacheEleName(info.IDC, fence["ef_name"])
			return flag
		}else{
			cacheInToubleEles(info.IDC,"")
		}
	}
	return flag

}
