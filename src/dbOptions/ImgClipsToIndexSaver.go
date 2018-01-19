package dbOptions

import (
	"fmt"
)

/**
	保存大图中某个小图的 index
 */
func ImgClipsToIndexSaver(index []byte, dbId uint8, mainImgId []byte, which uint8)  {
	imgIdent := GetImgClipIdent(dbId, mainImgId, which)
	err := InitImgClipsIndexDB().WriteTo([]byte(imgIdent),index)
	if nil != err{
		fmt.Println(err)
	}
}

/**
	指定大图中的某个小图，读取其属于哪些大图
 */
func ImgClipsToIndexReader(dbId uint8, mainImgId []byte, which uint8) []byte {
	imgIdent := GetImgClipIdent(dbId, mainImgId, which)
	return InitImgClipsIndexDB().ReadFor([]byte(imgIdent))
}