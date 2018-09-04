package yxtTool

import (
	"bytes"
	"net"
	"static"
	"strconv"
	"strings"
	"time"
)

/**
获取本地ip
*/
func GetLoaclIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addrss := range addrs {
		if ipnet, ok := addrss.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

/**
数组转16 进制字符串
*/
func ByteToHexString(b []byte) string {
	var hexStr = ""
	for _, i := range b {
		s := strconv.FormatInt(int64(i&0xff), 16)
		if len(s) < 2 {
			s = MergeString("0", s)
		}
		hexStr = MergeString(hexStr, s)
		hexStr = strings.ToUpper(hexStr)
	}
	return hexStr
}

/**
数组转BCD码
*/
func ByteToBCDString(b []byte) string {
	var hexStr = ""
	for _, i := range b {
		s := strconv.FormatInt(int64(i), 16)
		if len(s) < 2 {
			s = MergeString("0", s)
		}
		hexStr = MergeString(hexStr, s)

		hexStr = strings.ToUpper(hexStr)
	}
	return hexStr
}

/**
数组合并
*/
func Merge(b1 []byte, b2 []byte) (b []byte) {
	count := len(b1) + len(b2)
	b = make([]byte, count)
	index := copy(b, b1)
	copy(b[index:], b2)
	return

}

func MergeString(strs ...string) string {
	b := bytes.NewBufferString("")
	for _, value := range strs {
		b.WriteString(value)
	}
	return b.String()
}

/*
两个字节 转 uint16
*/
func GetUint16(bytes []byte) uint16 {
	return (uint16(bytes[0]) << 8) | uint16(bytes[1])
}

/*
四个字节 转 uint32
*/
func GetUint32(bytes []byte) uint32 {
	return (uint32(bytes[0]) << 24) | (uint32(bytes[1]) << 16) | (uint32(bytes[2]) << 8) | uint32(bytes[3])
}

func HexStringToByte(hexStr string) (bs []byte) {
	s_len := len(hexStr)
	bs = make([]byte, s_len/2)
	if s_len%2 == 0 {
		for i := 0; i < s_len/2; i++ {
			hex := hexStr[2*i : 2*i+2]
			b, _ := strconv.ParseInt(hex, 16, 32)
			bs[i] = byte(b & 0xff)
		}
	} else {
		bs = nil

	}

	return

}

func FormatUtcTime(bytes []byte) string {
	utcTime := GetUint32(bytes)
	d := time.Unix(int64(utcTime), 0)
	return d.Format(static.TimeLayOut1)
}

func ParseLonglan(bytes []byte) float64 {
	du := bytes[0]
	fen := float64(bytes[1]) + float64(bytes[2])/100 + float64(bytes[3])/10000
	return float64(du) + fen/60

}

func FloatToString(f float64, n int) string {
	return strconv.FormatFloat(f, 'f', n, 64)

}
func StringToFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
func FormatFloat(f float64, n int) float64 {
	fStr := FloatToString(f, n)
	return StringToFloat(fStr)

}
