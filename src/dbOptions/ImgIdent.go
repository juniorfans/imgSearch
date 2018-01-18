package dbOptions

import (
	"fmt"
)

/**
	ident 是图片的 identify, 比 imgid 更加精确地表示了一个图片存储在哪个库, imgId 是多少
 */

//获得大图 imdId 的 ident: dbId+imgId
func GetImgIdent(dbId uint8, imgId []byte) []byte {
	ret := make([]byte, len(imgId)+1)
	ret[0] = byte(dbId)
	copy(ret[1:], imgId)
	return ret
}

/**
	获得 clip identify
	编码格式：
	dbid			1字节
	threadId		1字节
	imgSeqNo		3字节(千万级别)
	which			1字节
 */
func GetImgClipIdent(dbId uint8, imgId []byte, which uint8) []byte {
	ret := make([]byte, 6)
	ret[0] = byte(dbId)
	ret[1] = imgId[0]
	ret[5] = byte(which)

	if 4 != len(imgId){
		fmt.Println("imgid is not 4 bytes: ", len(imgId), " -- ", string(ParseImgKeyToPlainTxt(imgId)))
		return nil
	}


	imgSeqNo := BytesToInt32(imgId[1:])

	seqBytes := Int32ToBytes(imgSeqNo)
	copy(ret[2:], seqBytes[1:])	//2,3,4
	return ret
}

func ParseImgClipIdentBytes(clipIdent []byte) (dbId uint8, imgId []byte, which uint8)  {
	if 6 != len(clipIdent){
		fmt.Println("parse img clip ident error, length is not 6: ", len(clipIdent))
		return
	}
	dbId = uint8(clipIdent[0])
	//threadId := clipIdent[1]
	imgId = make([]byte, 4)
	copy(imgId, clipIdent[1:5])
	which = uint8(clipIdent[5])
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

