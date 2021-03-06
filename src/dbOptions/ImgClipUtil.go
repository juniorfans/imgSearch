package dbOptions

import (
	"strings"
	"strconv"
	"fmt"
	"imgOptions"
	"bufio"
	"os"
	"config"
	"imgIndex"
	"github.com/Comdex/imgo"
	"util"
)

func SaveAllClipsOfImgs()  {
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
			SaveAllClipsAsJpgOf("E:/gen/clips/" , ImgIndex.FormatImgKey([]byte(imgKey)), []int{-1} ,-1)
		}
		imgDB.CloseDB()
	}
}

func PrintClipIndexFromClipIndexToIndent(dbId uint8)  {
	indexToClipDB := InitIndexToClipDB(dbId)
	iter := indexToClipDB.DBPtr.NewIterator(nil, &indexToClipDB.ReadOptions)
	iter.First()
	fmt.Println("clip index length is: ",len(iter.Key()))
}

func PrintClipIndex()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string
	var dbId, which uint8
	var mainImgId string

	for {
		fmt.Print("input dbId_mainImgId_which: ")
		fmt.Fscan(stdin, &input)
		inputs := strings.Split(input, "_")
		t, _ := strconv.Atoi(inputs[0])
		dbId = uint8(t)

		mainImgId = inputs[1]

		t, _ = strconv.Atoi(inputs[2])
		which = uint8(t)

		indexBytes := GetImgClipIndexFromClipIdent(dbId, ImgIndex.FormatImgKey([]byte(mainImgId)), which)
		fmt.Print(input," -- ")
		fileUtil.PrintBytes(indexBytes)
		fmt.Println()
		//fileUtil.PrintBytes(indexBytes)
	}
}

func PrintClipIndexInYCBCR()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string
	var dbId, which uint8
	var mainImgId string

	for {
		fmt.Print("input dbId_mainImgId_which: ")
		fmt.Fscan(stdin, &input)
		inputs := strings.Split(input, "_")
		t, _ := strconv.Atoi(inputs[0])
		dbId = uint8(t)

		mainImgId = inputs[1]

		t, _ = strconv.Atoi(inputs[2])
		which = uint8(t)

		//clip index 是 rgb 三个通道的字节. 所以是 3 的倍数
		indexBytes := GetImgClipIndexFromClipIdent(dbId, ImgIndex.FormatImgKey([]byte(mainImgId)), which)
		if len(indexBytes) % 3 != 0{
			fmt.Println("error, clip index length is not multiple of 3")
		}
		fmt.Print(input," -- ")
		nsize := len(indexBytes) / 3
		for i:=0;i < nsize ;i++  {
			PrintYCBCR(indexBytes[i*3:(i+1)*3])
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

func SaveSpecifiedClip()  {
	var dbId uint8
	var imgId string
	var which uint8

	for{
		fmt.Print("input dbId, imgId, which: ")

		stdin := bufio.NewReader(os.Stdin)
		fmt.Fscan(stdin, &dbId, &imgId, &which)
		imgDB := PickImgDB(dbId)
		imgKey := ImgIndex.FormatImgKey([]byte(imgId))
		indexes := GetDBIndexOfClips(imgDB , imgKey, []int{-1} ,-1)
		SaveClipsAsJpg("E:/gen/search/", indexes[which])

		clipIndexBytes := GetImgClipIndexFromClipIdent(dbId, imgKey, which)
		fmt.Print("clip index: ")
		fileUtil.PrintBytes(clipIndexBytes)

		branIndexes := ImgIndex.ClipIndexBranch(clipIndexBytes)
		fmt.Println("branches: ")
		for _, branIndex := range branIndexes{
			fileUtil.PrintBytes(branIndex)
		}
	}
}

func MarkClipIndexOnImg(imgDB *DBConfig)  {
	stdin := bufio.NewReader(os.Stdin)
	var input string

	for {
		fmt.Println("input img id or ids split by -")
		fmt.Fscan(stdin, &input)

		clipConfig := config.GetClipConfigById(0)
		imgKeyArray := strings.Split(input, "-")
		for _, imgKey := range imgKeyArray {
			SaveAllClipsAsJpgOf("E:/gen/clips/" , ImgIndex.FormatImgKey([]byte(imgKey)), clipConfig.ClipOffsets ,10)
		}
		imgDB.CloseDB()
	}
}

func SaveClipAsJpgFromIndexValue(value []byte, dir string)  {
	os.MkdirAll(dir, 0777)

	clipInfos := ImgIndex.ParseImgClipIdentListBytes(value)
	for _,clipInfo := range clipInfos{
		SaveAClipAsJpg(0,dir, clipInfo.DbId, clipInfo.ImgKey, clipInfo.Which)
	}
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


func SaveAllClipsAsJpgOf(dir string, mainImgkey []byte, offsetOfClip[] int, indexLength int){
	dbConfig := GetImgDBWhichPicked()
	if nil == dbConfig{
		fmt.Println("open imgdb error, ")
	}

	indexes := GetDBIndexOfClips(dbConfig,mainImgkey,offsetOfClip, indexLength)
	for _, index := range indexes{
		SaveClipsAsJpg(dir, index)
	}

	SaveMainImg(mainImgkey, dir)
}

func SaveAClipAsJpgFromClipIdent(dir string, clipIdent []byte)  {
	os.MkdirAll(dir, 0777)
	SaveAClipAsJpg(0, dir, clipIdent[0], clipIdent[1:ImgIndex.IMG_CLIP_IDENT_LENGTH-1], clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1])
}

func SaveAClipAsJpg(clipConfigId uint8, dir string, dbId uint8, mainImgkey []byte, which uint8){
	clipName := strconv.Itoa(int(dbId)) + "_" + string(ImgIndex.ParseImgKeyToPlainTxt(mainImgkey)) + "_" + strconv.Itoa(int(which))+".jpg"
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
	mainImgName := ImgIndex.ParseImgKeyToPlainTxt(indexData.KeyOfMainImg)
	clipName := strconv.Itoa(int(indexData.DBIdOfMainImg)) + "_" + string(mainImgName) + "_" +strconv.Itoa(int(indexData.Which))+".jpg"
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


func SaveClipsAsJpgWithName(dir, name string, indexData ImgIndex.SubImgIndex) {
	mainImgName := ImgIndex.ParseImgKeyToPlainTxt(indexData.KeyOfMainImg)
	clipName := strconv.Itoa(int(indexData.DBIdOfMainImg)) + "_" + string(mainImgName) + "_" +strconv.Itoa(int(indexData.Which))+"_"+name+".jpg"
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