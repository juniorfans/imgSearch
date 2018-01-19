package dbOptions

import (
	"fmt"
	"config"
	"bufio"
	"os"
	"strconv"
)

/**
	计算一张大图中的小图的协同关系
 */
func ExposeCalCollaborat(dbId uint8, imgId []byte)  {
	//取得大图的所有小图的 id
	clipConfig := config.GetClipConfigById(0)
	imgDB := PickImgDB(dbId)
	srcData := imgDB.ReadFor(imgId)
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,imgId,clipConfig.ClipOffsets, clipConfig.ClipLengh)
	if 0 == len(indexes){
		fmt.Println("save clips to db for ", string(imgId), " failed")
		return
	}
	for i:=0;i < len(indexes);i ++{
		i_index :=  indexes[i].GetFlatInfo()
		for j:=i+1;j < len(indexes) ;j++  {
			j_index := indexes[j].GetFlatInfo()

			fmt.Println("-------------------------------------------------")
			fmt.Println(indexes[i].Which, " -- ", indexes[j].Which)
			FindTwoClipsSameMainImgs(i_index ,j_index)
		}
	}
}



/**
	计算一张大图中的小图的协同关系
 */
func ExposeCalCollaboratEx(dbId uint8, imgId []byte)  {
	//取得大图的所有小图的 id
	clipConfig := config.GetClipConfigById(0)
	imgDB := PickImgDB(dbId)
	srcData := imgDB.ReadFor(imgId)
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,imgId,clipConfig.ClipOffsets, clipConfig.ClipLengh)
	if 0 == len(indexes){
		fmt.Println("save clips to db for ", string(imgId), " failed")
		return
	}
	for i:=0;i < len(indexes);i ++{
		i_index :=  indexes[i].GetFlatInfo()
		for j:=i+1;j < len(indexes) ;j++  {
			j_index := indexes[j].GetFlatInfo()

			fmt.Println("-------------------------------------------------")
			fmt.Println(indexes[i].Which, " -- ", indexes[j].Which)
			FindTwoClipsSameMainImgs(i_index ,j_index)
		}
	}
}

/**
	计算一张大图中的小图的协同关系
 */
func ExposeCalCollaboratWith(dbId uint8, imgId []byte, whichl, whichr uint8)  {
	//取得大图的所有小图的 id
	clipConfig := config.GetClipConfigById(0)
	imgDB := PickImgDB(dbId)
	srcData := imgDB.ReadFor(imgId)
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,imgId,clipConfig.ClipOffsets, clipConfig.ClipLengh)
	if 0 == len(indexes){
		fmt.Println("save clips to db for ", string(imgId), " failed")
		return
	}

	i_index :=  indexes[whichl].GetFlatInfo()

	j_index := indexes[whichr].GetFlatInfo()

	fmt.Println("-------------------------------------------------")
	fmt.Println(indexes[int(whichl)].Which, " -- ", indexes[int(whichr)].Which)
	FindTwoClipsSameMainImgs(i_index ,j_index)
}


/**
	计算一张大图中的小图的协同关系
 */
func ExposeCalCollaboratWithEx(dbId uint8, imgId []byte, whichl, whichr uint8)  {
	i_index := ImgClipsToIndexReader(dbId,imgId,whichl)
	j_index := ImgClipsToIndexReader(dbId,imgId,whichr)
	fmt.Println("-------------------------------------------------")
	fmt.Println(whichl, " -- ", whichr)
	FindTwoClipsSameMainImgs(i_index ,j_index)
}

/**
	计算两个子图 left, right 所在大图的交集
	left, right 是子图的 index
 */
func FindTwoClipsSameMainImgs(left, right []byte)  {

	clipDB := InitImgClipsReverseIndexDB()

	lv := clipDB.ReadFor(left)	//left 在哪些大图中出现过
	rv := clipDB.ReadFor(right)

	lvlist := ParseImgClipIdentListBytesToStrings(lv)
	rvlist := ParseImgClipIdentListBytesToStrings(rv)

	if nil == lvlist || nil == rvlist{
		fmt.Println("lvlist or rvlist is null")
		return
	}

	fmt.Println("left is in images: ", lvlist)
	fmt.Println("right is in images: ", rvlist)

	//lvlist 和 vlist 中需要过滤出相同的照片
	lvlist = DeleteSameMainImg(lvlist)

	rvlist = DeleteSameMainImg(rvlist)
	res := make(map[string]int)
	for _,lvc := range lvlist{
		res[lvc] ++
		fmt.Println(lvc)
	}
	fmt.Println("----------------------------------")
	for _,rvc := range rvlist{
		res[rvc] ++
		fmt.Println(rvc)
	}

	for id,sup := range res{
		if sup > 1{
			dbId, imgId := ParseImgIden(id)
			fmt.Println("dbId: ", dbId, ", imgId: ", string(imgId), "sup: ", sup)
		}
	}
}



func DeleteSameMainImg(imgClipIdents []string) []string {
	filter := make(map[string][]byte)
	ret := make([]string, len(imgClipIdents))
	realCount := 0

	var clipInfo ClipIdentInfo
	for _, clipIdent := range imgClipIdents {
		if 0 == len(clipIdent){
			fmt.Println("empty ident")
			continue
		}
		if !clipInfo.ParseFromIdenString(clipIdent){

			continue
		}

		imgDB := PickImgDB(clipInfo.dbId)
		imgBytes := imgDB.ReadFor(clipInfo.imgKey)

		if 0 == len(imgBytes){
			fmt.Println("img bytes of ",strconv.Itoa(int(clipInfo.dbId)), "-", string(ParseImgKeyToPlainTxt(clipInfo.imgKey)), " is empty")
			continue
		}else{
			fmt.Println("img bytes of ", strconv.Itoa(int(clipInfo.dbId)), "-", string(ParseImgKeyToPlainTxt(clipInfo.imgKey)), " is : ", len(imgBytes))
		}

		imgIndex := GetImgIndexBySrcData(imgBytes)
		imgIndexStr := string(imgIndex)
		if nil == filter[imgIndexStr]{
			filter[imgIndexStr] = clipInfo.imgKey
			ret[realCount]= clipIdent
			realCount ++
		}else{
			//abort
			//fmt.Println("abort: ", string(imgId))
		}
	}
	return ret[0:realCount]
}

func FindClipMainImg(dbId uint8, imgId []byte, which uint8)  {
	clipIdent := GetImgClipIdent(dbId, imgId, which)
	clipIndex := InitImgClipsIndexDB().ReadFor(clipIdent)
	if nil == clipIndex{
		fmt.Println("can't find clip index in clip_to_index db by clip ident")
		return
	}
	imgsInfo := InitImgClipsReverseIndexDB().ReadFor(clipIndex)
	if nil == imgsInfo{
		fmt.Println("can't find clip's img info in clip_reverse_index db by clip index")
		return
	}
	valueList := ParseImgClipIdentListBytesToStrings(imgsInfo)
	fmt.Println(valueList)
}

func TestFindClipMainImg()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId , which uint8
	var imgKey string
	for{
		fmt.Print("input dbId, imgKey, which: ")
		fmt.Fscan(stdin, &dbId, &imgKey, & which)
		FindClipMainImg(dbId, FormatImgKey([]byte(imgKey)), which)
	}
}