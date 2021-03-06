package dbOptions

import (
	"strconv"
	"util"
	"fmt"
	"imgCache"
	"imgIndex"
	"bytes"
)

var theStatIndexDBCache *imgCache.MyConcurrentMap

//非线程安全
func resetStatIndexDBQueryCache(cmap *imgCache.MyConcurrentMap) *imgCache.MyConcurrentMap {
	ret := theStatIndexDBCache
	theStatIndexDBCache = cmap
	return ret
}

/**
	计算哪些大图中联合出现了 imgKey 中的多个子图, imgKey 不包含在内
 */
func occInImgsEx(dbId uint8, imgKey []byte) (occedImgIndex *imgCache.MyMap, clipIndexAndIdents[][]byte, allStatBranchesIndex[] [][]byte ){
	curImgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	curImgIdent[0] = byte(dbId)
	copy(curImgIdent[1:], imgKey)
	curImgIndex := InitImgToIndexDB(uint8(dbId)).ReadFor(curImgIdent)

	clipIndexAndIdents = QueryClipIndexesAttachIdentFor(dbId, imgKey)

	if nil == clipIndexAndIdents {
		fmt.Println("can't find clip indexes: ", string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)))
		return
	}

	cachedDBSeeker := NewMultyDBReader(GetInitedClipStatIndexToIdentDB(), theStatIndexDBCache)

	//各个子图的 imgIndexContainer 容器: 每个容器表示某些 img index 与相应的子图有关系(即子图也出现在这些 img 中)
	occedImgIndex = imgCache.NewMyMap(true)

	allStatBranchesIndex = make([] [][]byte, len(clipIndexAndIdents))

	curClipOccIn := imgCache.NewMyMap(false)


	groupLen := ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH

	cacheHitCounts := 0
	for i, clipIndexAndIdent := range clipIndexAndIdents {

		clipIndex := clipIndexAndIdent[: ImgIndex.CLIP_INDEX_BYTES_LEN]
		curStatBranches := ImgIndex.ClipStatIndexBranch(clipIndex)
		allStatBranchesIndex[i] = curStatBranches

		//对于当前的第 i 个子图, 已经判断过了哪些子图与它相似. 再次遇到这些子图时可跳过计算
		dealedClipIndexes := imgCache.NewMyMap(false)

		//计算所有与当前子图相似的子图出现在哪此大图中
		for _,branch := range curStatBranches{

			clipIndexAndIdentsSet, cacheHit := cachedDBSeeker.ReadFor(branch)
			if cacheHit{
				cacheHitCounts ++
			}

			if 0 == len(clipIndexAndIdentsSet){
				continue
			}

			for _,clipIndexAndIdents := range clipIndexAndIdentsSet {

				for l:=0;l < len(clipIndexAndIdents);l += groupLen {
					curIndexAndIdent := clipIndexAndIdents[l:l + groupLen]
					curIndex := curIndexAndIdent[:ImgIndex.CLIP_INDEX_BYTES_LEN]

					if dealedClipIndexes.Contains(curIndex){
						continue
					}

					curIdent := curIndexAndIdent[ImgIndex.CLIP_INDEX_BYTES_LEN:]
					dealedClipIndexes.Put(curIndex, nil)
					if isSameClip(clipIndex, curIndex){
						curClipOccIn.Put(curIdent, nil)
					}
				}
			}
		}

		clipIdents := curClipOccIn.KeySet()
		if 0 == len(clipIdents){
			continue
		}

		curClipOccedImgIndexes := getImgIndexFromClipIdents(clipIdents)
		imgIndexes := curClipOccedImgIndexes.KeySet()
		if len(clipIdents) > 100{
			fmt.Println("OOPS, more than 100 same clip: ", len(clipIdents))
		}
		for _,imgIndex := range imgIndexes{
			//当前子图出现在下面的 img 中. 为了唯一性表示，使用 img index 作为键去表示
			interfaceClipIdent := curClipOccedImgIndexes.Get(imgIndex)
			if 1 != len(interfaceClipIdent){
				continue
			}
			clipIdent := interfaceClipIdent[0].([]byte)

			imgIdent := clipIdent[0:5]

			//当前图跳过
			if bytes.Equal(curImgIdent, imgIdent){
				continue
			}

			//与当前大图是同一张图
			if bytes.Equal(curImgIndex, imgIndex){
				continue
			}

			occedImgIndex.Put(imgIndex, uint8(i))	//第 i 个子图出现在 imgIndex 所指示的大图中
		}
		curClipOccIn.Clear()
	}

	//fmt.Println("---------------------- cache hit count: ", cacheHitCounts)
	return
}



//func splitToGroupsForEachIn(datas [][]byte, groupLen int) [][]byte{
//	if len(datas) == 0{
//		return nil
//	}
//
//	res := imgCache.NewMyMap(false)
//	for _,data := range datas{
//		for i:=0;i < len(data);i+=groupLen{
//			res.Put(data[i: i+groupLen], nil)
//		}
//	}
//	return res.KeySet()
//}


func isSameClip(leftClipIndex, rightClipIndex []byte) bool {
	//return ImgIndex.TheclipSearchConf.Delta_Eul_square > fileUtil.CalEulSquare(left, right)
	//直接依次比较元素的差值, 超过一定范围则认为不相等
	tolerant := uint8(2 * ImgIndex.TheclipSearchConf.Delta_Eul)
	for i, le := range leftClipIndex {
		ri := rightClipIndex[i]
		if le > ri && le-ri > tolerant{
			return false
		}

		if ri > le && ri-le > tolerant{
			return false
		}
	}
	return true

	//是否需要进一步的改进: 若 left[:i] 与 right[0:i] 的均值相差较大则认为不相似
}

func getImgIndexFromClipIdents(clipIdents [] []byte) *imgCache.MyMap {
	cmap := imgCache.NewMyMap(false)
	for _,clipIdent := range clipIdents{
		imgIndexDBId := clipIdent[0]
		imgIdent := clipIdent[0:5]
		imgIndex := InitImgToIndexDB(uint8(imgIndexDBId)).ReadFor(imgIdent)
		if len(imgIndex) == 0{
			continue
		}

		cmap.Put(imgIndex, clipIdent)
	}
	return cmap
}


/*
	以 dbId 库中的 imgKey 为对象，找出 imgKey 中哪些子图共同出现在其它大图中
*/
func SearchCoordinateForClipEx(dbId uint8, imgKey []byte) (whichesGroupAndCount *imgCache.MyMap, clipIndexAndIdents [][]byte, allStatIndex [] [][]byte ) {
	imgName := strconv.Itoa(int(dbId)) + "-" + string(ImgIndex.ParseImgKeyToPlainTxt(imgKey))
	//计算哪些大图中联合出现了 imgKey 中的多个子图. 注意 imgKey 不包含在内
	occedImgIndex, clipIndexAndIdents, allStatIndex := occInImgsEx(dbId, imgKey)

	if nil == occedImgIndex || 0 == occedImgIndex.KeyCount(){
		return
	}

	var resWhiches [][]uint8

	motherImgIndexes := occedImgIndex.KeySet()

	for _,imgIndex := range motherImgIndexes {
		interfaceWhiches := occedImgIndex.Get(imgIndex)
		if 2 > len(interfaceWhiches){
			continue
		}
		whiches := make([]uint8, len(interfaceWhiches))
		for i,which := range interfaceWhiches{
			whiches[i] = which.(uint8)
		}
		whiches = fileUtil.RemoveDupplicatedBytes(whiches)
		if 2 > len(whiches){
			continue
		}

		fileUtil.BytesSort(whiches)

		resWhiches = append(resWhiches, whiches)

	}
	if 0 == len(resWhiches){
		return
	}

	if len(resWhiches) > 1{
		fmt.Println("okay, find len(resWhiches) > 1: ", len(resWhiches), ", ",imgName)
	}

	//校准次数. 注意校验不能在 statCoordinateResult 中边统计边校准：只能最终校准.
	//原因在于: 设 1,3,4 同时出现在 A, B 图中，3,4 同时出现在 A,B,C 中则
	whichesGroupAndCount = statCoordinateResult(resWhiches)
	whichesGroups := whichesGroupAndCount.KeySet()
	for _,whiches := range whichesGroups{

		interfaceCounts := whichesGroupAndCount.Get(whiches)
		if 1 == len(interfaceCounts){
			countExclusiveCurrentImg := interfaceCounts[0].(int)
			whichesGroupAndCount.Put(whiches, countExclusiveCurrentImg + 1)
		}
	}


	//打印
	if len(whichesGroups) > 0{
		showStr := imgName + " : "
		for _,whiches := range whichesGroups{
			showStr += "["
			for _,which := range whiches{
				showStr += strconv.Itoa(int(which)) + ","
			}
			showStr += "]"
			interfaceCounts := whichesGroupAndCount.Get(whiches)
			if 1 == len(interfaceCounts){
				showStr += "-" + strconv.Itoa(interfaceCounts[0].(int)) + " | "
			}
		}
		fmt.Println(showStr)
	}

	return
}

func CalClipStatBranchIndexes(clipIdent []byte) [][]byte {
	dbId ,_,_:= ImgIndex.ParseAImgClipIdentBytes(clipIdent)
	clipIndex := InitClipToIndexDB(dbId).ReadFor(clipIdent)
	return ImgIndex.ClipStatIndexBranch(clipIndex)
}

func PrintClipStatBranchIndexes(clipIdent []byte)  {
	indexes := CalClipStatBranchIndexes(clipIdent)
	fmt.Println("stat indexes for ", clipIdent)
	for _,index := range indexes{
		fileUtil.PrintBytes(index)
	}
}

func TestClipStatBranchIndeses() {
	clipIdent1 := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIdent1[0] = 2
	copy(clipIdent1[1:], ImgIndex.FormatImgKey([]byte("F0000067")))
	clipIdent1[5] = 7

	clipIdent2 := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIdent2[0] = 2
	copy(clipIdent2[1:], ImgIndex.FormatImgKey([]byte("A0000000")))
	clipIdent2[5] = 3

	clipIdent3 := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIdent3[0] = 2
	copy(clipIdent3[1:], ImgIndex.FormatImgKey([]byte("E0000150")))
	clipIdent3[5] = 0

	PrintClipStatBranchIndexes(clipIdent1)
	PrintClipStatBranchIndexes(clipIdent2)
	PrintClipStatBranchIndexes(clipIdent3)
}

func TestStatIndexValue()  {

	/*
	referClipIdents := [] []byte{[]byte{2,70,0,0,67,7}, []byte{2,65,0,0,0,3}, []byte{2,69,0,0,150,0} }
	indexBytes := []byte{222, 57}
	clipIdentList := InitClipStatIndexToIdentsDB(2).ReadFor(indexBytes)

	if 0 == len(clipIdentList) || len(clipIdentList) % ImgIndex.IMG_CLIP_IDENT_LENGTH != 0{
		fmt.Println("error")
		return
	}

	clipIdents := make([][]byte, len(clipIdentList) / ImgIndex.IMG_CLIP_IDENT_LENGTH)
	ci := 0
	for i:=0;i < len(clipIdentList); i += ImgIndex.IMG_CLIP_IDENT_LENGTH{
		clipIdents[ci] = fileUtil.CopyBytesTo(clipIdentList[i: i + ImgIndex.IMG_CLIP_IDENT_LENGTH])
		ci ++
	}


	for _,clipIdent := range clipIdents{
	//	fileUtil.PrintBytes(clipIdent)
		for i,refer := range referClipIdents{
			if bytes.Equal(clipIdent, refer){
				fmt.Println("contain ", i)
			}
		}
	}

*/

	InitStatIndexToClipDB(2)
	occInImgsEx(2, []byte{65,0,0,0})

}

