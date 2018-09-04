package ServierBean

import (
	"math"
	"strconv"
	"strings"
)

const EARTH_RADIUS = 6378.137

type Point struct {
	/**
	经度
	*/
	Longitude float64

	/**
	纬度
	*/
	Latitude float64
}

func (p *Point) Lon() (lon string) {
	lon = strconv.FormatFloat(p.Longitude, 'f', -1, 64)
	return
}
func (p *Point) Lat() (lat string) {
	lat = strconv.FormatFloat(p.Latitude, 'f', -1, 64)
	return
}

/**
判断点在围栏内
*/
func (p *Point) PtInEleFence(fence map[string]string) (flag bool) {
	ef_type := fence["ef_type"]
	ef_range := fence["ef_range"]
	points := strings.Split(ef_range, ";")
	if ef_type == "矩形区域" || ef_type == "多边形区域" || ef_type == "行政区域" {
		//如果取得的ele_fence type为2 则是多边形区域
		var ps = make([]*Point, len(points))
		for i, item := range points {
			pxy := strings.Split(item, ",")
			lon, _ := strconv.ParseFloat(pxy[0], 64)
			lat, _ := strconv.ParseFloat(pxy[1], 64)
			ps[i] = &Point{lon, lat}
		}
		if len(ps) == 2 {
			//如果为两个点则是矩形区域点坐标为 左下，右上
			p4 := ps[0]
			p2 := ps[1]
			p1 := &Point{p4.Longitude, p2.Latitude}
			p3 := &Point{p2.Longitude, p4.Latitude}
			ps = []*Point{p1, p2, p3, p4}
		}
		flag = p.ptInPolygon(ps)
	} else if ef_type == "圆形区域" {
		r, _ := strconv.ParseFloat(points[0], 64)
		pxy := strings.Split(points[1], ",")
		lon, _ := strconv.ParseFloat(pxy[0], 64)
		lat, _ := strconv.ParseFloat(pxy[1], 64)
		center := &Point{lon, lat}
		flag = p.ptInCircle(r, center)
	}
	return
}

/*
点在圆内
*/
func (p *Point) ptInCircle(radius float64, center *Point) bool {
	distanse := getDistance(p, center) * 1000
	if distanse <= radius {
		//Console.WriteLine("在圆内");
		return true
	} else {
		//Console.WriteLine("在圆外");
		//Console.WriteLine("超出:" + (distanse - radius) + "公里");
		return false
	}
}
func rad(d float64) float64 {
	return d * math.Pi / 180.0
}

/**
计算地球两个点之间的距离 返回结果为KM
*/
func getDistance(p1 *Point, p2 *Point) float64 {
	radLat1 := rad(p1.Latitude)
	radLat2 := rad(p2.Latitude)
	a := radLat1 - radLat2
	b := rad(p1.Longitude) - rad(p2.Longitude)
	s := 2 * math.Asin(math.Sqrt(math.Pow(math.Sin(a/2), 2)+math.Cos(radLat1)*math.Cos(radLat2)*math.Pow(math.Sin(b/2), 2)))
	s = s * EARTH_RADIUS
	s = float64(int(s*10000)) / 10000
	return s
}

/**
判断点多边形内 ljh
*/
func (p *Point) ptInPolygon(dss []*Point) bool {
	x := p.Longitude
	y := p.Latitude
	i := 0
	polySides := len(dss)
	j := polySides - 1
	oddNodes := 0 // 0 false 1 true    go 不支持 bool 异或运算

	for i = 0; i < polySides; i++ {
		polyXi := dss[i].Longitude
		polyYi := dss[i].Latitude  //纬度
		polyXj := dss[j].Longitude //经度
		polyYj := dss[j].Latitude  //纬度
		if (polyYi < y && polyYj >= y || polyYj < y && polyYi >= y) && (polyXi <= x || polyXj <= x) {
			temp := polyXi+(y-polyYi)/(polyYj-polyYi)*(polyXj-polyXi) < x
			if temp {
				oddNodes ^= 1
			} else {
				oddNodes ^= 0
			}
		}
		j = i
	}
	return oddNodes == 1
}
