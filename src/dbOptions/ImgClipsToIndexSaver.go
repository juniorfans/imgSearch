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
	err := InitClipToIndexDB(dbId).WriteTo([]byte(imgIdent),index)
	if nil != err{
		fmt.Println(err)
	}
}

/**
	保存大图中某个小图的 index
 */
func ImgClipsToIndexBatchSaver(dbId uint8,batch *leveldb.Batch)  {
	InitClipToIndexDB(dbId).WriteBatchTo(batch)
}

/**
	指定大图中的某个小图，读取它的 clip index
 */
func GetImgClipIndexFromClipIdent(dbId uint8, mainImgId []byte, which uint8) []byte {
	clipIdent := ImgIndex.GetImgClipIdent(dbId, mainImgId, which)
	fileUtil.PrintBytes(clipIdent)
	return InitClipToIndexDB(dbId).ReadFor([]byte(clipIdent))
}