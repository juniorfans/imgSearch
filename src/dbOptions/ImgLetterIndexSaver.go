package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
	"imgOptions"
	"imgIndex"
	"strconv"
	"config"
	"github.com/Comdex/imgo"
	"os"
)

/**
	[TODO 这里直接使用了 subIndex[0] ，只使用了一个]

 */

func SaveLetterIndexForImg(imgDB *DBConfig,imgKey []byte)  {
	letterConfig := config.GetLetterConfigById(0)

	srcData , err := imgDB.DBPtr.Get(imgKey, &imgDB.ReadOptions)
	if err == leveldb.ErrNotFound{
		fmt.Println("can't get the img: ", string(imgKey))
		return
	}
	data := ImgOptions.FromImageFlatBytesToStructBytes(srcData)
	subIndex := ImgIndex.GetLetterIndexFor(letterConfig, data, imgKey)
	if nil == subIndex{
		return
	}

	indexBytes := ImgIndex.GetFlatIndexBytesFrom([]ImgIndex.SubImgIndex{subIndex[0]})
	if nil == indexBytes{
		fmt.Println("GetFlatIndexBytesFrom error")
		return
	}

	imgLetterDB := InitImgLetterDB()
	if nil == imgLetterDB{
		fmt.Println("open img letter db error")
		return
	}

	oldValue,err := imgLetterDB.DBPtr.Get(indexBytes, &imgLetterDB.ReadOptions)
	var newValue string
	if err == leveldb.ErrNotFound{
		newValue = string(oldValue)
	}else{
		newValue = string(oldValue) + "-" + strconv.Itoa(imgDB.Id) + "-" + string(imgKey)
	}
	imgLetterDB.DBPtr.Put(indexBytes,[]byte(newValue), &imgLetterDB.WriteOptions)
}


func SaveLetterIndexAsJpgForImg(imgDB *DBConfig,imgKey []byte, dir string)  {
	letterConfig := config.GetLetterConfigById(0)

	srcData , err := imgDB.DBPtr.Get(imgKey, &imgDB.ReadOptions)
	if err == leveldb.ErrNotFound{
		fmt.Println("can't get the img: ", string(imgKey))
		return
	}
	imgData := ImgOptions.FromImageFlatBytesToStructBytes(srcData)
	subIndex := ImgIndex.GetLetterIndexFor(letterConfig, imgData, imgKey)
	if nil == subIndex{
		return
	}

	data := imgo.New3DSlice(letterConfig.Height, letterConfig.Width, 4)

	for _, indexUnit := range subIndex[0].IndexUnits{
		ImgIndex.IndexDataApplyIntoSubImg(data,indexUnit)
	}

	fileName := dir + "/" + string(imgKey) + "_letter_.jpg"
	fmt.Println("gen name: " , dir + "/" + fileName)
	os.MkdirAll(dir, 0777)
	imgo.SaveAsJPEG( fileName ,data,100)

}