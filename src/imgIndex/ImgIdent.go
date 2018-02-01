package ImgIndex

import (
	"fmt"
	"config"
	"util"
)

type ClipIdentInfo struct {
	DbId   uint8
	ImgKey []byte
	Which  uint8
}

func (this *ClipIdentInfo) GetThreadId() uint8 {
	return uint8(this.ImgKey[0])
}
func (this *ClipIdentInfo) ToBytes() []byte {
	return GetImgClipIdent(this.DbId, this.ImgKey, this.Which)
}
func (this *ClipIdentInfo) ToString() string {
	return string(GetImgClipIdent(this.DbId, this.ImgKey, this.Which))
}

func NewClipIdentInfo(clipIdentBytes []byte) *ClipIdentInfo {
	ret := ClipIdentInfo{}
	dbId, imgKey, which := ParseAImgClipIdentBytes(clipIdentBytes)
	ret.DbId = uint8(dbId)
	ret.ImgKey = imgKey
	ret.Which = uint8(which)
	return &ret

}

var IMG_KEY_LENGTH = 4				//img key 的长度
var IMG_IDENT_LENGTH = IMG_KEY_LENGTH+1	//img ident 的长度, 由 img key 加上 dbId 构成
var IMG_CLIP_IDENT_LENGTH = IMG_KEY_LENGTH+2	//img clip ident 的长度, 由 img key 再加上 dbId, which 构成


func GetImgKey(threadId uint8, seqNo int) []byte {
	ret := make([]byte, 4)	//3字节表达的 uint 范围是0 ~ 16777215

	ret[0]=byte(config.ThreadIdToByte[int(threadId)])

	seqBytes := Int32ToBytes(seqNo)
	copy(ret[1:], seqBytes[1:])
	return ret
}

/**
	ident 是图片的 identify, 比 imgid 更加精确地表示了一个图片存储在哪个库, imgId 是多少

	目前是 6 个字节
 */

//获得大图 imdId 的 ident: dbId+imgId
func GetImgIdent(dbId uint8, imgKey []byte) []byte {
	if len(imgKey) != IMG_KEY_LENGTH{
		fmt.Println("img key is not ", IMG_KEY_LENGTH)
		return nil
	}
	ret := make([]byte, IMG_IDENT_LENGTH)
	ret[0] = byte(dbId)
	copy(ret[1:], imgKey)
	return ret
}

/**
	获得 clip identify
	编码格式：
	dbid			1字节

	<下面两个字节刚好就是 img key>
	threadId		1字节
	imgSeqNo		3字节(千万级别)

	which			1字节
 */
func GetImgClipIdent(dbId uint8, imgId []byte, which uint8) []byte {
	ret := make([]byte, IMG_CLIP_IDENT_LENGTH)
	ret[0] = byte(dbId)
	copy(ret[1:], imgId)
	ret[IMG_CLIP_IDENT_LENGTH-1] = byte(which)
	return ret
}

func FromClipIdentToImgIdent(clipIdent []byte) []byte{
	ret := make([]byte, IMG_IDENT_LENGTH)
	ci:=0
	ci += copy(ret[ci:], clipIdent[0:IMG_IDENT_LENGTH])
	return ret
}

func FromClipIdentsToStrings(clipIdentListBytes []byte) [] string{
	fmt.Println("clip belong to how many imgs: ", len(clipIdentListBytes) / IMG_CLIP_IDENT_LENGTH)
	infos := ParseImgClipIdentListBytes(clipIdentListBytes)
	if nil == infos{
		return nil
	}
	ret := make([]string, len(infos))
	for i, info := range infos {
		fmt.Println(info.DbId, "-", string(ParseImgKeyToPlainTxt(info.ImgKey)), "-", info.Which)
		ret[i] = info.ToString()
	}
	return ret
}


func ParseImgClipIdentListBytes(clipIdentListBytes []byte) [] ClipIdentInfo{
	if len(clipIdentListBytes) % IMG_CLIP_IDENT_LENGTH != 0{
		fmt.Println("clip ident list bytes lenth is not multiple of ", IMG_CLIP_IDENT_LENGTH, ": ", len(clipIdentListBytes))
		fileUtil.PrintBytes(clipIdentListBytes)
		return nil
	}
	nsize := len(clipIdentListBytes)/IMG_CLIP_IDENT_LENGTH
	ret := make([]ClipIdentInfo, nsize)

	for i:=0;i < nsize;i ++{
		dbId, imgKey, which := ParseAImgClipIdentBytes(clipIdentListBytes[i*IMG_CLIP_IDENT_LENGTH : (i+1)*IMG_CLIP_IDENT_LENGTH])
		ret[i].DbId = dbId
		ret[i].ImgKey = imgKey
		ret[i].Which = which
	}

	return ret
}

func ParseAImgClipIdentBytes(clipIdent []byte) (dbId uint8, imgId []byte, which uint8)  {
	if IMG_CLIP_IDENT_LENGTH != len(clipIdent){
		fmt.Println("parse img clip ident error, length is not 6: ", len(clipIdent))
		return
	}
	dbId = uint8(clipIdent[0])
	//threadId := clipIdent[1]
	imgId = make([]byte, IMG_KEY_LENGTH)
	copy(imgId, clipIdent[1:IMG_CLIP_IDENT_LENGTH-1])
	which = uint8(clipIdent[IMG_CLIP_IDENT_LENGTH-1])
	return
}


func ParseImgIdenBytes(ibytes []byte) (dbId uint8, imgId []byte) {
	dbId = uint8(ibytes[0])

	imgId = make([]byte, len(ibytes)-1)
	copy(imgId, ibytes[1:])
	return
}


func ParseImgIden(ident string) (dbId uint8, imgId []byte) {
	return ParseImgIdenBytes([]byte(ident))
}

