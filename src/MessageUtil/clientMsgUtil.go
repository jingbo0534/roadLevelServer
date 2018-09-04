package MessageUtil

import (
	"math"
	"strconv"
	"strings"
)

/**
params: clientMessage 不含帧头和帧尾的数组
msg_id 消息Id
sim_b id号码
msg_num

*/

func getSimNo(b []byte) (bcdStr string) {
	bcdStr = ""
	for _, i := range b {
		s := strconv.FormatInt(int64(i&0xff), 10)
		bcdStr += s
		bcdStr = strings.ToUpper(bcdStr)
	}
	return
}

const PI float64 = math.Pi                     //圆周率
const earthR float64 = 6378245.0               //地球半径 单位m
const earthEE float64 = 0.00669342162296594323 //系数EE

//地球坐标系转换到火星坐标系
func WgsTogcj(gcjLng float64, gcjLat float64) []string {

	lnglat := make([]string, 2)
	dLat := transFormLat(gcjLng-105.0, gcjLat-35.0)
	dLng := transFormLng(gcjLng-105.0, gcjLat-35.0)
	radLat := gcjLat / 180.0 * PI
	magic := math.Sin(radLat)
	magic = 1 - earthEE*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((earthR * (1 - earthEE)) / (magic * sqrtMagic) * PI)
	dLng = (dLng * 180.0) / (earthR / sqrtMagic * math.Cos(radLat) * PI)
	gcjLng += dLng
	gcjLat += dLat
	lnglat[0] = strconv.FormatFloat(gcjLng, 'f', 6, 64)
	lnglat[1] = strconv.FormatFloat(gcjLat, 'f', 6, 64)
	return lnglat
}

//转换成纬度
func transFormLat(x1 float64, y1 float64) float64 {
	ret := -100.0 + 2.0*x1 + 3.0*y1 + 0.2*y1*y1 + 0.1*x1*y1 + 0.2*math.Sqrt(math.Abs(x1))
	ret += (20.0*math.Sin(6.0*x1*PI) + 20.0*math.Sin(2.0*x1*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y1*PI) + 40.0*math.Sin(y1/3.0*PI)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y1/12.0*PI) + 320*math.Sin(y1*PI/30.0)) * 2.0 / 3.0
	return ret
}

//转换成经度
func transFormLng(x2 float64, y2 float64) float64 {

	ret := 300.0 + x2 + 2.0*y2 + 0.1*x2*x2 + 0.1*x2*y2 + 0.1*math.Sqrt(math.Abs(x2))
	ret += (20.0*math.Sin(6.0*x2*PI) + 20.0*math.Sin(2.0*x2*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x2*PI) + 40.0*math.Sin(x2/3.0*PI)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x2/12.0*PI) + 300.0*math.Sin(x2/30.0*PI)) * 2.0 / 3.0
	return ret
}
