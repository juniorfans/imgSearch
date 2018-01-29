package dbOptions

import (
	"fmt"
	"config"
	"imgIndex"
	"util"
	"imgCache"
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
		i_index :=  indexes[i].GetIndexBytesIn3Chanel()
		for j:=i+1;j < len(indexes) ;j++  {
			j_index := indexes[j].GetIndexBytesIn3Chanel()

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
		i_index :=  indexes[i].GetIndexBytesIn3Chanel()
		for j:=i+1;j < len(indexes) ;j++  {
			j_index := indexes[j].GetIndexBytesIn3Chanel()

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

	i_index :=  indexes[whichl].GetIndexBytesIn3Chanel()

	j_index := indexes[whichr].GetIndexBytesIn3Chanel()

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

	if nil == i_index{
		fmt.Println("read clip index nil: ", dbId, ", ", string(ParseImgKeyToPlainTxt(imgId)), ",", whichl)
	}
	if nil == j_index{
		fmt.Println("read clip index nil: ", dbId, ", ", string(ParseImgKeyToPlainTxt(imgId)), ",", whichr)
	}

	fmt.Println("-------------------------------------------------")
	fmt.Println(whichl, " -- ", whichr)
	FindTwoClipsSameMainImgs(i_index ,j_index)
}

/**
	计算两个子图 left, right 所在大图的交集
	left, right 是子图的 index
 */
func FindTwoClipsSameMainImgs(left, right []byte)  {

	indexToClip := InitIndexToClipDB()

	leftBranches := ImgIndex.ClipIndexBranch(4,10, left)
	rightBranches := ImgIndex.ClipIndexBranch(4,10, right)
	var lvlist, rvlist []byte

	{
		for _,lb := range leftBranches{
			curlist := indexToClip.ReadFor(lb)	//left 在哪些大图中出现过
			fileUtil.MergeBytesTo(&lvlist, &curlist)
		}
	}

	{
		for _,rb := range rightBranches{
			curList := indexToClip.ReadFor(rb)	//left 在哪些大图中出现过
			fileUtil.MergeBytesTo(&rvlist, &curList)
		}
	}

	if 0 == len(lvlist){
		fmt.Println("lvlist is null")
		return
	}
	if 0 == len(rvlist){
		fmt.Println("rvlist is null")
		return
	}


	//indexBytes to indexIdent
	lmap := RemoveDuplicate(lvlist)
	rmap := RemoveDuplicate(rvlist)

	FindSameImg(lmap, rmap)

}



func RemoveDuplicate(imgClipIdents []byte) *imgCache.MyMap {
	imgIndexToIdent := imgCache.NewMyMap(false)

	for i:=0;i < len(imgClipIdents);i += IMG_CLIP_IDENT_LENGTH{
		imgClipIdent := FromClipIdentToImgIdent(imgClipIdents[i:i+IMG_CLIP_IDENT_LENGTH])
		imgIdent := imgClipIdent[0:IMG_CLIP_IDENT_LENGTH-1]
		imgIndexBytes := InitImgToIndexDB().ReadFor(imgIdent)
		if nil == imgIndexBytes{
			fmt.Println("img index nil: ", string(ParseImgKeyToPlainTxt(imgIdent[1:])))
		}
		imgIndexToIdent.Put(imgIndexBytes, imgIdent)
	}
	return imgIndexToIdent
}

func FindSameImg(left, right *imgCache.MyMap) *imgCache.MyMap {

	//img index bytes -- img ident
	statMap := imgCache.NewMyMap(true)

	combineVisitor := &combineVisitor{}
	left.Visit(combineVisitor, -1, []interface{}{statMap})
	right.Visit(combineVisitor, -1, []interface{}{statMap})

	resultMap := imgCache.NewMyMap(false)
	removeDuplicateVisitor := &removeDuplicateVisitor{}
	statMap.Visit(removeDuplicateVisitor, -1, []interface{}{resultMap})

	return resultMap
}


//-------------------------------------------------------------------------------------------
type combineVisitor struct {

}

func (this *combineVisitor) Visit(imgIndexBytes []byte, imgIdent []interface{}, otherParams []interface{}) bool {
	if 1 != len(otherParams){
		fmt.Println("NoNameVisitor other params not 1")
		return false
	}

	statMap := otherParams[0].(*imgCache.MyMap)
	statMap.Put(imgIndexBytes, imgIdent[0])

	return true
}


type removeDuplicateVisitor struct {

}

func (this *removeDuplicateVisitor) Visit(imgIndexBytes []byte, imgIdents []interface{}, otherParams []interface{}) bool {
	if 1 != len(otherParams){
		fmt.Println("NoNameVisitor other params not 1")
		return false
	}

	resultMap := otherParams[0].(*imgCache.MyMap)

	count := len(imgIdents)
	if 1 < count{
		resultMap.Put(imgIndexBytes, count)
		imgIdent := imgIdents[0].([]byte)
		fmt.Println("------------ ", count)
		fmt.Println(ParseImgIdentToPlainTxt(imgIdent)," : ", count)

	}

	return true
}