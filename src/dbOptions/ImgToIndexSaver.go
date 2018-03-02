package dbOptions

import "github.com/syndtr/goleveldb/leveldb"

func ImgToIndexSaver(dbId uint8, imgId []byte, indexBytes[]byte){
	imgToIndexDB := InitImgToIndexDB(dbId)
	imgToIndexDB.WriteTo(imgId, indexBytes)

}


func ImgToIndexBatchSaver(dbId uint8, batch *leveldb.Batch){
	imgToIndexDB := InitImgToIndexDB(dbId)
	imgToIndexDB.WriteBatchTo(batch)

}