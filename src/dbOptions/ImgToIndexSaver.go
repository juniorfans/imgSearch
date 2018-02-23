package dbOptions

import "github.com/syndtr/goleveldb/leveldb"

func ImgToIndexSaver(dbId uint8, imgId []byte, indexBytes[]byte){
	imgToIndexDB := InitMuImgToIndexDB(dbId)
	imgToIndexDB.WriteTo(imgId, indexBytes)

}


func ImgToIndexBatchSaver(dbId uint8, batch *leveldb.Batch){
	imgToIndexDB := InitMuImgToIndexDB(dbId)
	imgToIndexDB.WriteBatchTo(batch)

}