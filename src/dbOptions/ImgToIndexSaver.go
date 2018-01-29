package dbOptions

import "github.com/syndtr/goleveldb/leveldb"

func ImgToIndexSaver(imgId []byte, indexBytes[]byte){
	imgToIndexDB := InitImgToIndexDB()
	imgToIndexDB.WriteTo(imgId, indexBytes)

}


func ImgToIndexBatchSaver(batch *leveldb.Batch){
	imgToIndexDB := InitImgToIndexDB()
	imgToIndexDB.WriteBatchTo(batch)

}