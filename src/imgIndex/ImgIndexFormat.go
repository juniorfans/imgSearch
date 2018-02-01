package ImgIndex

import (
	"fmt"
	"strconv"
	"bytes"
	"encoding/binary"
)

func MakeSurePlainImgIdIsOk(plainImgId []byte) []byte {
	if 0 == len(plainImgId){
		fmt.Println("error, key is empty")
		return nil
	}
	if 8 < len(plainImgId){
		fmt.Println("error, key length is more than 8")
		return nil
	}
	if 8 == len(plainImgId){
		return plainImgId
	}

	ret := make([]byte, 8)	//7 位数字可保存百万级别图片, 再加一位字母线程标识，总共 8 位
	ci := 0

	ret[0]= plainImgId[0]
	ci ++

	for i:=len(plainImgId);i < 8;i ++{	//填充 8-len(oldKey) 个 '0'
		ret[ci]=byte('0')
		ci ++
	}

	ci += copy(ret[ci:], plainImgId[1:])

	if ci != 8 {
		fmt.Println("new key cal error, ci: ", ci, ", not 8")
		return nil
	}
	return ret
}


/**
	输入：一个字节线程标识，后面字节为 seqNo
	使用 4 个字节作为 imgId， 表达千万级别图片
	线程标识		1字节
	imgSeqNo	3字节
 */
func FormatImgKey(oldKey []byte) []byte {
	if IMG_KEY_LENGTH == len(oldKey){
		return oldKey
	}

	if 0 == len(oldKey){
		fmt.Println("error, key is empty")
		return nil
	}
	if 8 < len(oldKey){
		fmt.Println("error, key length is more than 8")
		return nil
	}

	ret := make([]byte, IMG_KEY_LENGTH)	//3字节表达的 uint 范围是0 ~ 16777215

	ret[0]=oldKey[0]	//threadid

	seqNo, err := strconv.Atoi(string(oldKey[1:]))
	if nil != err{
		fmt.Println("parse imgid error: ", string(oldKey))
		return nil
	}
	seqBytes := Int32ToBytes(seqNo)
	copy(ret[1:], seqBytes[1:])

	return ret
}

func ParseImgKeyToPlainTxt(imgKey []byte) []byte {
	if 4 != len(imgKey){
		fmt.Println("imgKey format error, must be 4 bytes,but is ", len(imgKey), " : ", string(MakeSurePlainImgIdIsOk(imgKey)))
		return nil
	}
	ret := make([]byte, 8)
	ret[0] = imgKey[0]

	plainTxt := string(imgKey[0:1])
	plainTxt += strconv.Itoa(int(BytesToInt32(imgKey[1:])))
	return MakeSurePlainImgIdIsOk([]byte(plainTxt))
}

func ParseImgIdentToPlainTxt(imgIdent []byte) string {
	if IMG_IDENT_LENGTH != len(imgIdent){
		fmt.Println("imgKey format error, must be 4 bytes,but is ", len(imgIdent))
		return ""
	}

	plainTxt := string(ParseImgKeyToPlainTxt(imgIdent[1:]))

	return "" + strconv.Itoa(int(imgIdent[0])) + "-" + string(MakeSurePlainImgIdIsOk([]byte(plainTxt)))
}


//大端模式, 字节数组中高位是原值中的低位
func BytesToInt32(b []byte) int {
	if 4 > len(b){
		ret := 0
		ci := uint(0)
		factor := 0
		for i:=len(b)-1;i!=-1;i--{
			factor = 1 << (8*ci)
			ci ++

			ret += int(b[i])*factor
		}
		return ret
	}
	//下面的操作需要 b 的长度是 4
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

//大端模式
func Int32ToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian,tmp)
	return bytesBuffer.Bytes()

}

//imgId 是大端模式: 原值中低位在现在字节数组中的高位，注意顺序
//第 0 位是 threadName, 第 1~3 位是 imgKey
func ImgIdIncrement(imgId []byte) bool {
	if uint8(imgId[3]) < 255{
		imgId[3] = imgId[3]+1
	}else if uint8(imgId[2]) < 255{
		imgId[2] = imgId[2]+1
		imgId[3] = 0
	}else if uint8(imgId[1]) < 255{
		imgId[1] = imgId[1]+1
		imgId[2] = 0
		imgId[3] = 0
	}else{
		fmt.Println("imgId is too big, can't increment")
		return false
	}
	return true
}

//imgId 是大端模式: 原值中低位在现在字节数组中的高位，注意顺序
//第 0 位是 threadName, 第 1~3 位是 imgKey
func ImgIdDecrement(imgId []byte) bool {
	if uint8(imgId[3]) > 0{
		imgId[3] = imgId[3]-1
	}else if uint8(imgId[2]) > 0{
		imgId[2] = imgId[2]-1
		imgId[3] = 255
	}else if uint8(imgId[1]) > 0{
		imgId[1] = imgId[1]-1
		imgId[2] = 255
		imgId[3] = 255
	}else{
		fmt.Println("imgId is too small, can't decrement")
		return false
	}
	return true
}

//字节转换成整形, 小端模式
func BytesToUInt32(b []byte) uint32 {
	//bytesBuffer := bytes.NewBuffer(b)
	//var tmp int32
	//binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	//return int(tmp)

	if 1 == len(b){
		return uint32(b[0])
	}else if 2 == len(b){
		return uint32(b[1])*256 + uint32(b[0])
	}else if 3 == len(b){
		return uint32(b[2])*65536 + uint32(b[1])*256 + uint32(b[0])
	}else if 4 == len(b){
		return uint32(b[3])*16777216 + uint32(b[2])*65536 + uint32(b[1])*256 + uint32(b[0])
	}else{
		fmt.Println("BytesToUint32 error, out of bytes range: ", len(b))
		return 0
	}
}

//小端模式: 低位在低地址
func UInt32ToBytes(n uint32) []byte {
	//tmp := int32(n)
	//bytesBuffer := bytes.NewBuffer([]byte{})
	//binary.Write(bytesBuffer, binary.BigEndian,tmp)
	//return bytesBuffer.Bytes()
	if n <= 255{
		return []byte{byte(n)}
	}else if n <= 65535{
		return []byte{byte(n%256),byte(n/256)}
	}else if n <= 16777215{
		return []byte{byte(n%256), byte(n/256 % 256), byte(n/65536)}
	}else if n <= 4294967295{
		return []byte{byte(n%256), byte(n/256 % 256), byte(n/65536 % 256), byte(n/16777216)}
	}else{
		fmt.Println("Uint32ToBytes error, out of range: ", n)
		return nil
	}

}