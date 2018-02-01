package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"util"
	"imgIndex"
)

/**
	保存大图中某个小图的 index
 */
func ImgClipsToIndexSaver(index []byte, dbId uint8, mainImgId []byte, which uint8)  {
	imgIdent := ImgIndex.GetImgClipIdent(dbId, mainImgId, which)
	err := InitMuClipToIndexDb(dbId).WriteTo([]byte(imgIdent),index)
	if nil != err{
		fmt.Println(err)
	}
}

/**
	保存大图中某个小图的 index
 */
func ImgClipsToIndexBatchSaver(dbId uint8,batch *leveldb.Batch)  {
	InitMuClipToIndexDb(dbId).WriteBatchTo(batch)
}

/**
	指定大图中的某个小图，读取其属于哪些大图
 */
func GetImgClipIndexFromClipIdent(dbId uint8, mainImgId []byte, which uint8) []byte {
	imgIdent := ImgIndex.GetImgClipIdent(dbId, mainImgId, which)
	fileUtil.PrintBytes(imgIdent)
	return InitMuClipToIndexDb(dbId).ReadFor([]byte(imgIdent))
}