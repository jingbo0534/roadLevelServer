package yxtTool

/**
需要校验的数据进行校验 传入的数组含有校验位
*/
func DoCheckBytes(clientbytes []byte) (enCodeBytes []byte) {
	bLen := len(clientbytes) //长度 包含校验码
	enCodeBytes = make([]byte, bLen)
	copy(enCodeBytes, clientbytes)
	checkCode := enCodeBytes[0]
	for i := 1; i < bLen-1; i++ {
		checkCode = checkCode ^ enCodeBytes[i]
	}
	enCodeBytes[bLen-1] = checkCode
	return
}

/**
反转意数据部分
*/
func DeTransform(enByte []byte) (deByte []byte) {
	deByte = make([]byte, len(enByte)*2)
	var realLen = 0
	for i := 0; i < len(enByte); i++ {
		item := enByte[i]
		if item == 0x7d {
			if enByte[i+1] == 0x01 {
				deByte[realLen] = 0x7d
				realLen++
				i++
				continue
			}
			if enByte[i+1] == 0x02 {
				deByte[realLen] = 0x7e
				realLen++
				i++
				continue
			}
		}
		deByte[realLen] = item
		realLen++
	}
	deByte = deByte[0:realLen]
	return
}

/**
转意数据部分
*/
func EnTransform(deByte []byte) (enByte []byte) {
	enByte = make([]byte, len(deByte)*2)
	var realLen = 0
	for _, item := range deByte {
		if item == 0x7d {
			enByte[realLen] = 0x7d
			realLen += 1
			enByte[realLen] = 0x01
			realLen += 1
		} else if item == 0x7e {
			enByte[realLen] = 0x7d
			realLen += 1
			enByte[realLen] = 0x02
			realLen += 1
		} else {
			enByte[realLen] = item
			realLen += 1
		}
	}
	enByte = enByte[0:realLen]
	return
}
