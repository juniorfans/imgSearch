package dbOptions

func ImgToIndexSaver(imgId []byte, indexBytes[]byte){
	imgToIndexDB := InitImgToIndexDB()
	imgToIndexDB.WriteTo(imgId, indexBytes)

}
