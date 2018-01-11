package dbOptions

import (
	"fmt"
	"github.com/Comdex/imgo"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"image/jpeg"
	"io"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"strings"
	"imgIndex"
)



var caclFinished chan int

func threadFunc(threadId int, iter iterator.Iterator, count int, offsetOfClip, indexLength int)  {
	groupName := config.ThreadIdToName[threadId]
	iter.Seek([]byte(groupName))
	if !iter.Valid(){
		fmt.Println("thread ", threadId, " iterator is invalid")
		return
	}
	for iter.Valid(){
		if 0 != strings.Index(string(iter.Key()), groupName){
			break
		}
		fmt.Println("thread ", threadId, " dealing: ", string(iter.Key()))

		SaveAllClipsToDBOf(iter.Value(), iter.Key(), offsetOfClip, indexLength)

		iter.Next()
		count --
		if count <= 0{
			break
		}
	}

	caclFinished <- threadId
}

func Calc(count int, offsetOfClip, indexLength int) {
	cores := 8
	caclFinished = make(chan int, cores)

	imgDB := InitImgDB()
	InitImgClipsDB()

	for i:=0;i != cores;i++{
		go threadFunc(i, imgDB.DBPtr.NewIterator(nil,&imgDB.ReadOptions),count, offsetOfClip, indexLength)
	}

	for i:=0;i < cores;i ++{
		threadId := <- caclFinished
		fmt.Println("thread ", threadId ," finished")
	}
	fmt.Println("All finished ~")
}

func SaveAllClipsToDBOf(srcData []byte, mainImgkey []byte, offsetOfClip, indexLength int){

	//获得 mainImgKey 的各个切图的索引数据
	indexes := GetDBIndexOfClipsBySrcData(srcData,mainImgkey,offsetOfClip, indexLength)

	imgClipDB := InitImgClipsDB()
	if nil == imgClipDB{
		fmt.Println("open img_clip db error, ")
		return
	}

	//保存各个索引数据
	for _, index := range indexes{
		SaveClipsToDB(imgClipDB, index)
	}

	//ReadValues(imgClipDB.DBPtr, 100)
}

func SaveAllClipsAsJpgOf(mainImgkey []byte, offsetOfClip, indexLength int){
	dbConfig := InitImgDB()
	if nil == dbConfig{
		fmt.Println("open imgdb error, ")
	}

	indexes := GetDBIndexOfClips(dbConfig,mainImgkey,offsetOfClip, indexLength)
	for _, index := range indexes{
		SaveClipsAsImg(index)
	}
}

func SaveClipsAsImg(indexData ImgIndex.ImgClipIndex) {
	mainImgName := string(indexData.KeyOfMainImg)
	clipName := mainImgName+strconv.Itoa(int(indexData.Which))+".jpg"
	clipConfig := config.GetClipConfigById(indexData.ClipConfigId)

	//复原索引到图片数据中，若索引数据只是原图片数据的部分(理应如此)，则恢复出来的图片只有部分的图像
	data := imgo.New3DSlice(clipConfig.SmallPicWidth, clipConfig.SmallPicHeight, 4)
	indexes := indexData.IndexUnits
	for _, index := range indexes{
		recoverClip(data, clipConfig.SmallPicHeight, clipConfig.SmallPicWidth, index)
	}

	fmt.Println("gen name: " , "E:/gen/" + clipName)
	imgo.SaveAsJPEG("E:/gen/" + clipName,data,100)
}

func SaveClipsToDB(clipDBConfig *ImgDBConfig, indexData ImgIndex.ImgClipIndex) {
	index := indexData.GetFlatInfo()

	exsitsMainImgKey,err := clipDBConfig.DBPtr.Get(index, &clipDBConfig.ReadOptions)
	if err != leveldb.ErrNotFound{
		newMainKeys := string(exsitsMainImgKey) + "-"+string(indexData.KeyOfMainImg)
		clipDBConfig.DBPtr.Put(index,[]byte(newMainKeys), &clipDBConfig.WriteOptions)
		return
	}

	clipDBConfig.DBPtr.Put(index,indexData.KeyOfMainImg,&clipDBConfig.WriteOptions)
}

/**
	将 indexUnit 指示的一个索引单元复原到 data 中
 */
func recoverClip(data [][][]uint8, height, width int, indexUnit ImgIndex.IndexUnit)  {
	count := 0

	var pcs  []ImgIndex.PointColor = indexUnit.Index
	realCount := 0
	realCountLimit := len(pcs)
	for j:=0;j < height;j++  {
		for i:=0;i < width;i++  {
			if count >= indexUnit.Offset{
				if realCount>= realCountLimit{
					return
				}
				var colors [4]uint8 = [4]uint8(pcs[realCount])
				data[j][i][0] = colors[0]
				data[j][i][1] = colors[1]
				data[j][i][2] = colors[2]
				data[j][i][3] = colors[3]
				realCount ++
			}
			count ++
		}
	}
	return
}

/**
	获得 fileName 图像中的小图
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据

	offsetOfClip 为负数则置为0，大于边界则返回 nil
	indexLength 过大或者为负数则返回从 offsetOfClip 开始的所有数据
 */
func GetDBIndexOfClips(dbConfig *ImgDBConfig,mainImgkey []byte, offsetOfClip, indexLength int) []ImgIndex.ImgClipIndex {
	srcData,err := dbConfig.DBPtr.Get(mainImgkey, &dbConfig.ReadOptions)
	if err == leveldb.ErrNotFound{
		fmt.Println("not found image key: ", string(mainImgkey), err)
		return nil
	}
	return GetDBIndexOfClipsBySrcData(srcData, mainImgkey, offsetOfClip, indexLength)
}


/**
	获得 fileName 图像中的小图
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据

	offsetOfClip 为负数则置为0，大于边界则返回 nil
	indexLength 过大或者为负数则返回从 offsetOfClip 开始的所有数据
 */
func GetDBIndexOfClipsBySrcData(srcData []byte,mainImgkey []byte, offsetOfClip, indexLength int) []ImgIndex.ImgClipIndex {
	var reader io.Reader = bytes.NewReader([]byte(srcData))

	image, err := jpeg.Decode(reader)
	if nil != err{
		fmt.Println("not jpeg data: ", string(mainImgkey), err)
		return nil
	}
	data, err := imgo.Read(image)
	if nil != err{
		fmt.Println("read jpeg data error: ", string(mainImgkey), err)
		return nil
	}

	return ImgIndex.GetClipsIndexOfImg(data, mainImgkey, offsetOfClip, indexLength)

}
