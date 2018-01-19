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

	lastDealedKey,curCount := GetThreadLastDealedKey(InitImgClipsReverseIndexDB(), dbIndex, threadId)

	iter.First()

	if nil != lastDealedKey {
		iter.Seek(lastDealedKey)
		if iter.Valid(){
			iter.Next()
		}
	}

	if !iter.Valid(){
		fmt.Println("thread ", threadId, " iterator is invalid or none data to deal")
		return
	}

	baseCount := 0
	failedCount := 0

	curKey := make([]byte, 8)

	for {
		if !iter.Valid(){
			iter.Prev()
			lastDealedKey = iter.Key()
			break
		}

		if baseCount!=0 && 0 == baseCount % 1000{
			fmt.Println("thread ", threadId, " dealing: ", baseCount)
		}

		copy(curKey, iter.Key())
		if !SaveAllClipsToDBOf(iter.Value(),dbIndex, curKey[0:len(iter.Key())], offsetOfClip, indexLength){
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

	fmt.Println("lastValue: threadId: ", threadId, " -- ",ParseImgKeyToPlainTxt(lastDealedKey))

	SetThreadLastDealedKey(InitImgClipsReverseIndexDB(),dbIndex, threadId, lastDealedKey, curCount)
	fmt.Println("thread ", threadId," dealed: ", baseCount ,", failedCount: ", failedCount)
	caclFinished <- threadId

	iter.Release()
}


func BeginImgClipSave(dbIndex uint8, count int, offsetOfClip []int, indexLength int) {
	cores := 16
	runtime.GOMAXPROCS(cores)

	caclFinished = make(chan int, cores)

	InitImgClipsReverseIndexDB()

	for i:=0;i < cores;i++{
		go threadFunc(dbIndex,i,count, offsetOfClip, indexLength)
	}

	for i:=0;i < cores;i ++{
		threadId := <- caclFinished
		fmt.Println("thread ", threadId ," finished")
	}

	RepairTotalSize(InitImgClipsReverseIndexDB())

	fmt.Println("All finished ~")
}


func SaveAllClipsToDBOf(srcData []byte,dbId uint8,  mainImgkey []byte, offsetOfClip []int, indexLength int) bool{

	//获得 mainImgKey 的各个切图的索引数据
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,mainImgkey,offsetOfClip, indexLength)
	if nil == indexes{
		fmt.Println("save clips to db for ", string(mainImgkey), " failed")
		return false
	}

	imgClipDB := InitImgClipsReverseIndexDB()
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
	clipName := strconv.Itoa(int(dbId)) + "_" + string(ParseImgKeyToPlainTxt(mainImgkey)) + "_" + strconv.Itoa(int(which))+".jpg"
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

/**
	将原 oldValue 与新的 clip value 合并, 支持 oldValue 为 nil
 */
func getValueForClipsKeyEx(oldValue []byte, indexData ImgIndex.SubImgIndex) []byte {
	if len(oldValue) % IMG_CLIP_IDENT_LENGTH != 0{
		fmt.Println("old clip ident length is not multy of ", IMG_CLIP_IDENT_LENGTH)
		return nil
	}

	if len(indexData.KeyOfMainImg) == 0{
		fmt.Println("fuck, real empty")
	}

	clipIdent := GetImgClipIdent(indexData.DBIdOfMainImg,indexData.KeyOfMainImg,indexData.Which)



	ret := make([]byte,len(oldValue)+IMG_CLIP_IDENT_LENGTH)
	ci := 0
	if 0 != len(oldValue){
		ci += copy(ret[ci:], oldValue)
	}
	ci += copy(ret[ci:], clipIdent)
	return ret
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
	index = EditClipIndex(index)

	oldValue, err := clipDBConfig.DBPtr.Get(index, &clipDBConfig.ReadOptions)
	if err == leveldb.ErrNotFound{
		oldValue = nil
	}
	clipValue := getValueForClipsKeyEx(oldValue, indexData)//getValueForClipsKey(oldValue, indexData)
	//fmt.Println("newMainImaKey: ", newMainKeys)
	clipDBConfig.DBPtr.Put(index,[]byte(clipValue), &clipDBConfig.WriteOptions)

	ImgClipsToIndexSaver(index, indexData.DBIdOfMainImg, indexData.KeyOfMainImg, indexData.Which)
}

func SaveClipAsJpgFromIndexValue(value []byte, dir string)  {
	os.MkdirAll(dir, 0777)

	clipInfos := ParseImgClipIdentListBytes(value)
	for _,clipInfo := range clipInfos{
		SaveAClipAsJpg(0,dir, clipInfo.dbId, clipInfo.imgKey, clipInfo.which)
	}
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
		fmt.Println("not found image key: ", ParseImgKeyToPlainTxt(mainImgkey), err)
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
		fmt.Println("read jpeg data error: ", ParseImgKeyToPlainTxt(mainImgkey))
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
			SaveAllClipsAsJpgOf("E:/gen/clips/" , ParseImgKeyToPlainTxt([]byte(imgKey)), clipConfig.ClipOffsets ,10)
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
			SaveAllClipsAsJpgOf("E:/gen/clips/" , ParseImgKeyToPlainTxt([]byte(imgKey)), []int{-1} ,-1)
		}
		imgDB.CloseDB()
	}
}

func PrintClipIndexBytes()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string
	var dbId, which uint8
	var mainImgId string
	InitImgClipsIndexDB()

	for {
		fmt.Print("input dbId_mainImgId_which: ")
		fmt.Fscan(stdin, &input)
		inputs := strings.Split(input, "_")
		t, _ := strconv.Atoi(inputs[0])
		dbId = uint8(t)

		mainImgId = inputs[1]

		t, _ = strconv.Atoi(inputs[2])
		which = uint8(t)

		indexBytes := ImgClipsToIndexReader(dbId, FormatImgKey([]byte(mainImgId)), which)
		fmt.Print(input," -- ")
		nsize := len(indexBytes) / 4
		for i:=0;i < nsize ;i++  {
			PrintYCBCR(indexBytes[i*4:(i+1)*4])
		}
		fmt.Println()
		//fileUtil.PrintBytes(indexBytes)
	}
}

func PrintYCBCR(rgba []uint8)  {
	ycbcr := make([]uint8, 3)
	ycbcr[0],ycbcr[1], ycbcr[2] = ImgOptions.TranRGBToYCBCR(rgba[0], rgba[1], rgba[2])
	fmt.Print(ycbcr[0]," ",ycbcr[1], " ", ycbcr[2],"| ")
}