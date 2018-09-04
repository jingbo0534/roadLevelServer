package static

const TimeLayOut1 = "2006-01-02 15:04:05"
const TimeLayOut2 = "060102150405"
const TimeLayOut3 = "2006-01-02"
const TimeLayOut4 = "15:04"
const RoadLevelFile="RoadLevelMake"
const RoadVoiceUrl="http://www.tsgcpt.cn/zqcw/sendVoice/sendVoiceTTS"
const AlarmVoiceUrl="http://www.tsgcpt.cn/zqcw/sendVoice/sendAlarmVoice"

var  IDCS = []string{"192500104500","192500104230","192500104730","192500104180","192500104680"}

func IDCIn(idc string)bool  {
	for _,v:=range IDCS{
		if v==idc{
			return true
		}
	}
	return true
}