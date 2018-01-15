package dbOptions

import (
	"fmt"
	"github.com/Comdex/imgo"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
	"imgIndex"
	"imgOptions"
	"bufio"
	"os"
	"runtime"
	"github.com/syndtr/goleveldb/leveldb/util"
)



var caclFinished chan int

func threadFunc(dbIndex uint8, threadId int, count int, offsetOfClip []int, indexLength int)  {
	srcDB := GetImgDBWhichPicked()
	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := srcDB.DBPtr.NewIterator(&region,&srcDB.ReadOptions)

	lastDealedKey,curCount := GetThreadLastDealedKey(InitImgClipsDB(), dbIndex, threadId)

	iter.First()

	if nil != lastDealedKey {
		iter.Seek(lastDealedKey)
		if iter.Valid(){
			iter.Next()
		}
	}

	if !iter.Valid(){
		fmt.Println("thread ", threadId, " iterator is invalid")
		return
	}

	baseCount := 0
	failedCount := 0
	for {
		if !iter.Valid(){
			iter.Prev()
			lastDealedKey = iter.Key()
			break
		}

		if baseCount!=0 && 0 == baseCount % 1000{
			fmt.Println("thread ", threadId, " dealing: ", baseCount)
		}

		if !SaveAllClipsToDBOf(iter.Value(),dbIndex, iter.Key(), offsetOfClip, indexLength){
			failedCount ++
			continue
		}

		iter.Next()

		baseCount ++
		count --
		if count <= 0{
			lastDealedKey = iter.Key()
			break
		}
	}

	curCount += baseCount

	fmt.Println("lastValue: threadId: ", threadId, " -- ",string(lastDealedKey))

	SetThreadLastDealedKey(InitImgClipsDB(),dbIndex, threadId, lastDealedKey, curCount)
	fmt.Println("thread ", threadId," dealed: ", baseCount ,", failedCount: ", failedCount)
	caclFinished <- threadId

	iter.Release()
}


func BeginImgClipSave(dbIndex uint8, count int, offsetOfClip []int, indexLength int) {
	cores := 8
	runtime.GOMAXPROCS(cores)

	caclFinished = make(chan int, cores)

	InitImgClipsDB()

	for i:=0;i < cores;i++{
		go threadFunc(dbIndex,i,count, offsetOfClip, indexLength)
	}

	for i:=0;i < cores;i ++{
		threadId := <- caclFinished
		fmt.Println("thread ", threadId ," finished")
	}

	RepairTotalSize(InitImgClipsDB())

	fmt.Println("All finished ~")
}


func SaveAllClipsToDBOf(srcData []byte,dbId uint8,  mainImgkey []byte, offsetOfClip []int, indexLength int) bool{

	//获得 mainImgKey 的各个切图的索引数据
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,mainImgkey,offsetOfClip, indexLength)
	if nil == indexes{
		fmt.Println("save clips to db for ", string(mainImgkey), " failed")
		return false
	}

	imgClipDB := InitImgClipsDB()
	if nil == imgClipDB{
		fmt.Println("open img_clip db error, ")
		return false
	}

	//保存各个索引数据
	for _, index := range indexes{
		SaveClipsToDB(imgClipDB, index)
	}

	//ReadValues(imgClipDB.DBPtr, 100)

	return true
}

func SaveAllClipsAsJpgOf(dir string, mainImgkey []byte, offsetOfClip[] int, indexLength int){
	dbConfig := GetImgDBWhichPicked()
	if nil == dbConfig{
		fmt.Println("open imgdb error, ")
	}

	indexes := GetDBIndexOfClips(dbConfig,mainImgkey,offsetOfClip, indexLength)
	for _, index := range indexes{
		SaveClipsAsJpg(dir, index)
	}

	SaveMainImg(string(mainImgkey), dir)
}

func SaveAClipAsJpg(clipConfigId uint8, dir string, dbId uint8, mainImgkey []byte, which uint8){
	mainImgName := string(mainImgkey)
	clipName := mainImgName+"_"+strconv.Itoa(int(which))+".jpg"
	clipConfig := config.GetClipConfigById(clipConfigId)

	//复原索引到图片数据中，若索引数据只是原图片数据的部分(理应如此)，则恢复出来的图片只有部分的图像
	data := imgo.New3DSlice(clipConfig.SmallPicHeight, clipConfig.SmallPicWidth, 4)

	clips := GetDBIndexOfClips(PickImgDB(dbId),mainImgkey,[]int{-1},-1) //取所有的
	if nil == clips{
		fmt.Println("clip data error")
		return
	}
	if len(clips)-1 <  int(which){
		fmt.Println("clip which is too big, all count is ", len(clips), ", which is ", which)
		return
	}

	indexes := clips[int(which)].IndexUnits
	for _, index := range indexes{
		//ImgIndex.IndexDataApplyIntoSubImg(data, clipConfig.SmallPicHeight, clipConfig.SmallPicWidth, index)
		ImgIndex.IndexDataApplyIntoSubImg(data, index)
	}


	fmt.Println("gen name: " , dir + "/" + clipName)
	imgo.SaveAsJPEG(dir + "/" + clipName,data,100)
}

func SaveClipsAsJpg(dir string, indexData ImgIndex.SubImgIndex) {
	mainImgName := string(indexData.KeyOfMainImg)
	clipName := mainImgName+strconv.Itoa(int(indexData.Which))+".jpg"
	clipConfig := config.GetClipConfigById(indexData.ConfigId)

	//复原索引到图片数据中，若索引数据只是原图片数据的部分(理应如此)，则恢复出来的图片只有部分的图像
	data := imgo.New3DSlice(clipConfig.SmallPicHeight, clipConfig.SmallPicWidth, 4)
	indexes := indexData.IndexUnits
	for _, index := range indexes{
		//ImgIndex.IndexDataApplyIntoSubImg(data, clipConfig.SmallPicHeight, clipConfig.SmallPicWidth, index)
		ImgIndex.IndexDataApplyIntoSubImg(data, index)
	}

	fmt.Println("gen name: " , dir + "/" + clipName)
	imgo.SaveAsJPEG(dir + "/" + clipName,data,100)
}



func getValueForClipsKey(oldValue []byte, indexData ImgIndex.SubImgIndex) string {

	mainImgKey := string(indexData.KeyOfMainImg)

	dbId := strconv.Itoa(int(GetImgDBWhichPicked().Id))

	var newMainKeys string
	if 0 != len(oldValue){
		exsitsMainImgKey := string(oldValue)

		dotPos := strings.LastIndex(exsitsMainImgKey, "-")
		if -1==dotPos{
			//格式不对，可能已经损坏，重新计数
			newMainKeys = mainImgKey + "-" + dbId + "-2" //当前 clip-key 所属的 main image 总共有 2 个，其中一个是 exsitsMainImgKey，所在的 imgdb 是 dbId
		}else{
			count , cerr := strconv.Atoi(exsitsMainImgKey[dotPos+1:])
			if nil == cerr{
				count ++
				newMainKeys = exsitsMainImgKey[0:dotPos+1] + strconv.Itoa(count)
			}else{
				newMainKeys = mainImgKey + "-" +dbId + "-2"
			}
		}
	}else{
		newMainKeys = mainImgKey + "-" +dbId + "-1"	//当前 clip-key 所属的 main image 总共有 1 个
	}
	return newMainKeys
}


func getValueForClipsKeyEx(oldValue []byte, indexData ImgIndex.SubImgIndex) []byte {
	newValue := ClipIndexValue{Which:indexData.Which,DbId:indexData.DBIdOfMainImg,ImgId:indexData.KeyOfMainImg}
	if 0 == len(oldValue){
		valueList := InitClipIndexValueList()
		valueList.AppendClipVlue(&newValue)
		valueList.Finish()
		ret:= valueList.ToBytes()
	//	fmt.Println(ret)
		return ret
	}else{
		//直接将当前的 indexValue 追加到后面即可，注意，别忘记分隔符了
		splitBytes := InitClipIndexValueList().Splits
		newValueBytes := newValue.ToBytes()
		ret := make([]byte, len(oldValue) + len(splitBytes) + len(newValueBytes))
		ci := 0
		ci += copy(ret[ci:], oldValue)
		ci += copy(ret[ci:], splitBytes)
		ci += copy(ret[ci:], newValueBytes)
	//	fmt.Println(ret)
		return ret
	}
}

func getValueForClipsKeyForTest(oldValue []byte, indexData ImgIndex.SubImgIndex) string {
	mainImgKey := string(indexData.KeyOfMainImg)
	dbId := strconv.Itoa(int(GetImgDBWhichPicked().Id))

	if 0 == len(oldValue){
		newMainKeys := string(mainImgKey) + "-" + dbId
		return newMainKeys
	}

	newMainKeys := string(oldValue) + "-" + string(mainImgKey) + "-" + dbId
	if strings.Count(newMainKeys, "-") > 10{
		fmt.Println(newMainKeys)
	}
	return newMainKeys
}

func SaveClipsToDB(clipDBConfig *DBConfig, indexData ImgIndex.SubImgIndex) {
	index := indexData.GetFlatInfo()
	oldValue, err := clipDBConfig.DBPtr.Get(index, &clipDBConfig.ReadOptions)
	if err == leveldb.ErrNotFound{
		oldValue = nil
	}
	newMainKeys := getValueForClipsKeyEx(oldValue, indexData)//getValueForClipsKey(oldValue, indexData)
	//fmt.Println("newMainImaKey: ", newMainKeys)
	clipDBConfig.DBPtr.Put(index,[]byte(newMainKeys), &clipDBConfig.WriteOptions)

	ImgClipsToIndexSaver(index, indexData.DBIdOfMainImg, indexData.KeyOfMainImg, indexData.Which)
}

func SaveClipAsJpgFromIndexValue(value []byte, dir string)  {
	os.MkdirAll(dir, 0777)
	indexValList := ParseClipIndeValues(value)
	dbId, mainImgId, which := indexValList.WhereCanFindClip()
	SaveAClipAsJpg(0,dir, dbId,mainImgId,which )
	PickImgDB(dbId).CloseDB()
}

/**
	获得 fileName 图像中的小图
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据

	offsetOfClip 为负数则置为0，大于边界则返回 nil
	indexLength 过大或者为负数则返回从 offsetOfClip 开始的所有数据
 */
func GetDBIndexOfClips(dbConfig *DBConfig,mainImgkey []byte, offsetOfClip []int, indexLength int) []ImgIndex.SubImgIndex {
	srcData,err := dbConfig.DBPtr.Get(mainImgkey, &dbConfig.ReadOptions)
	if err == leveldb.ErrNotFound{
		fmt.Println("not found image key: ", string(mainImgkey), err)
		return nil
	}
	return GetDBIndexOfClipsBySrcData(srcData, dbConfig.Id, mainImgkey, offsetOfClip, indexLength)
}


/**
	获得 fileName 图像中的小图
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据

	offsetOfClip 为负数则置为0，大于边界则返回 nil
	indexLength 过大或者为负数则返回从 offsetOfClip 开始的所有数据
 */
func GetDBIndexOfClipsBySrcData(srcData []byte, dbId uint8,mainImgkey []byte, offsetOfClip []int, indexLength int) []ImgIndex.SubImgIndex {

	data := ImgOptions.FromImageFlatBytesToStructBytes(srcData)

	if nil == data{
		fmt.Println("read jpeg data error: ", string(mainImgkey))
		return nil
	}

	return ImgIndex.GetClipsIndexOfImgEx(data, dbId, mainImgkey, offsetOfClip, indexLength)
}

func TestClipsIndexSaveToJpgFromImgDB()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbIndex uint8
	var input string

	for {
		fmt.Println("select a image db to deal: ")
		fmt.Fscan(stdin, &dbIndex)
		imgDB := PickImgDB(dbIndex)
		if nil == imgDB{
			fmt.Println("open db: ", dbIndex, " error")
			continue
		}

		fmt.Println("input img id or ids split by -")
		fmt.Fscan(stdin, &input)

		clipConfig := config.GetClipConfigById(0)
		imgKeyArray := strings.Split(input, "-")
		for _, imgKey := range imgKeyArray {
			SaveAllClipsAsJpgOf("E:/gen/clips/" , FormatImgKey([]byte(imgKey)), clipConfig.ClipOffsets ,10)
		}
		imgDB.CloseDB()
	}
}

func TestClipsSaveToJpgFromImgDB()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbIndex uint8
	var input string

	for {
		fmt.Println("select a image db to deal: ")
		fmt.Fscan(stdin, &dbIndex)
		imgDB := PickImgDB(dbIndex)
		if nil == imgDB{
			fmt.Println("open db: ", dbIndex, " error")
			continue
		}

		fmt.Println("input img id or ids split by -")
		fmt.Fscan(stdin, &input)

		imgKeyArray := strings.Split(input, "-")
		for _, imgKey := range imgKeyArray {
			SaveAllClipsAsJpgOf("E:/gen/clips/" , FormatImgKey([]byte(imgKey)), []int{-1} ,-1)
		}
		imgDB.CloseDB()
	}
}