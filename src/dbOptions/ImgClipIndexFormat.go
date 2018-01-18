package dbOptions

import (
	"fmt"
	"bytes"
	"strings"
	"strconv"
)

type ClipIndexValue struct {
	Which uint8	//当前 clip 子图是大图的第几个:0 开始
	DbId uint8	//当前 clip 子图所在的大图，所在的 db
	ImgId []byte	//大图的 id
}


type ClipIndexValueList struct {
	SplitsLength uint8 //第一个字节指示了 Splits 占用的长度
	// 8-SplitsLength 个填充字节
	Splits []byte	//每个 clip index value 之间分隔的字节序列
	ClipValues []ClipIndexValue	//各个 clip index value

	//不存储的数据
	TotalBytes int	//所有 value 加上分隔符所占的字节数
}

func InitClipIndexValueList() ClipIndexValueList {
	return ClipIndexValueList{SplitsLength:1, Splits:[]byte{byte('-')}, TotalBytes:0}
}

func (this *ClipIndexValue)GetLength() int {
	return 2 + len(this.ImgId)
}

func (this *ClipIndexValue) ToBytes() []byte {
	ret := make([]byte, this.GetLength())
	ret[0] = byte(this.Which)
	ret[1] = byte(this.DbId)
	copy(ret[2:], this.ImgId)
	return ret
}

func (this *ClipIndexValue)ParseFrom (clipValue []byte){
	this.Which=uint8(clipValue[0])
	this.DbId=uint8(clipValue[1])
	this.ImgId=make([]byte, len(clipValue)-2)
	copy(this.ImgId, clipValue[2:])
}

func (this *ClipIndexValue) IsEqualTo(other *ClipIndexValue) bool {
	return this.DbId==other.DbId && bytes.Equal(this.ImgId, other.ImgId)
}

func (this *ClipIndexValueList) AppendClipVlue(value *ClipIndexValue)  {
	this.ClipValues=append(this.ClipValues, *value)
	this.TotalBytes += value.GetLength()
}

func (this *ClipIndexValueList) Finish()  {
	//暂时什么也不做
}

/**
	clip index 对应的 value 的格式是
	- 头	（8 个字节）
		1 字节表示分隔串的长度
		7 个字节填充
	- 分隔串
	- 6 字节 clip ident 信息
	- 分隔串

 */
func (this *ClipIndexValueList) ToBytes() []byte {
	values := this.ClipValues

	ret := make([]byte, this.TotalBytes + 8 + int(this.SplitsLength)) //所有 value 长度加上分隔符再加上最前面的 8 个字节作为头
	ci := 0
	ci += copy(ret[ci:], []byte{byte(this.SplitsLength)})
	ci += copy(ret[ci:], []byte{0,0,0,0,0,0,0})	//暂无意义的填充字节
	ci += copy(ret[ci:], this.Splits)
	for _, value := range values{
		ci += copy(ret[ci:], value.ToBytes())
	}
	return ret
}

/**
	clipValueBytes 前 8 个字节是头
 */
func ParseClipIndexValues(clipValueBytes []byte) ClipIndexValueList{

	if nil == clipValueBytes || 8 > len(clipValueBytes) {
		return ClipIndexValueList{}
	}

	splitBytesLen := uint8(clipValueBytes[0])
	//notUsedBytes := clipValueBytes[1:8]

	splitBytes := clipValueBytes[8: 8+splitBytesLen]
	clipValueBytes = clipValueBytes[8+splitBytesLen:]

	valueStrs := strings.Split(string(clipValueBytes), string(splitBytes))
	if len(valueStrs) == 0{
		fmt.Println("parse error")
		return ClipIndexValueList{}
	}

	indexValueList := InitClipIndexValueList()
	retClipValues := make([]ClipIndexValue, len(valueStrs))
	indexValueList.ClipValues = retClipValues

	for i,valueStr := range valueStrs{
		curVaueBytes := []byte(valueStr)
		if 0 == len(curVaueBytes){
			continue;
		}
		retClipValues[i].ParseFrom(curVaueBytes)
		indexValueList.TotalBytes += len(curVaueBytes)
	}

	return indexValueList
}

/**
	返回 clip 所在的大图的信息
 */
func (this *ClipIndexValueList) WhereCanFindClip() (dbId uint8, mainImgId []byte, which uint8) {
	values := this.ClipValues
	for _, value := range values{
		mainImgId =(value.ImgId)
		dbId = value.DbId
		which = value.Which
		return
	}
	return
}

func (this *ClipIndexValueList) Print() {
	fmt.Println("--------------------------------------------------")
	fmt.Println(string(this.Splits))
	fmt.Println("totalBytes: ", this.TotalBytes)
	values := this.ClipValues
	for _, value := range values{
		fmt.Println("id: ", string(value.ImgId), ", db: ",strconv.Itoa(int(value.DbId)), ", which: ", value.Which)
	}
}



//返回 clip 所属的大图 ident
func GetMainImgIdentOfClips(vlists *ClipIndexValueList) []string {
	ret :=make([]string, len(vlists.ClipValues))
	for i,vlist :=range vlists.ClipValues {
		ret[i] = string(GetImgIdent(vlist.DbId, vlist.ImgId))
	}
	return ret
}
