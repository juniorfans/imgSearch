package dbOptions

import (
	"imgIndex"
	"util"
	"fmt"
	"config"
	"imgCache"
	"bytes"
	"strconv"
)

//allBranches 数组中，第 cur 个 clip branches 是否与前面的重复
func hasDuplicateClipWithBefore(allBranches [][][]byte, cur int) bool {
	if 0 == cur{
		return false
	}
	curBranches := allBranches[cur]
	for i:=0;i < cur;i ++{
		beforeBranches := allBranches[i]
		if isDuplicateClipBranchesIndex(curBranches, beforeBranches){
			return true
		}
	}
	return false
}

func isDuplicateClipBranchesIndex(lefts, rights [][]byte) bool {
	for _,left := range lefts{
		for _,right := range rights{
			if IsSameClipBranchIndex(left, right){
				return true
			}
		}
	}
	return false
}

/**
	计算哪些大图中联合出现了 imgKey 中的多个子图, imgKey 不包含在内
	返回值: occedImgIndex 键为大图的 index, 值为子图编号 which 的数组
	cachedImgIndexToIdent 缓存了 imgIndex 到 imgIdent 的对应关系，后面需要用到，缓存起来减少查询次数
	allBranchesIndex 各个子图的分支索引, 三维数组

	需要用到的库: imgIdentToIndex
 */
func occInImgs(dbId uint8, imgKey []byte) (occedImgIndex *imgCache.MyMap, cachedImgIndexToIdent *imgCache.MyMap, allBranchesIndex[] [][]byte ){
	curImgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	curImgIdent[0] = byte(dbId)
	copy(curImgIdent[1:], imgKey)
	curImgIndex := InitImgToIndexDB(uint8(dbId)).ReadFor(curImgIdent)

	//clipConfig := config.GetClipConfigById(0)
	//clipIndexes := GetDBIndexOfClips(PickImgDB(dbId) , imgKey, clipConfig.ClipOffsets, clipConfig.ClipLengh)
	clipIndexes := QueryClipIndexesFor(dbId, imgKey)
	if nil == clipIndexes{
		fmt.Println("can't find clip indexes: ", string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)))
		return
	}

//	clipIndexToIdentSeeker := NewMultyClipBIndexToIdentSeeker(GetInitedClipIndexToIdentDB())
	//发现一个现象，如果参考库过多，噪声会增加，这样会使得将本不是同主题的子图划分为一个主题。反而，若只使用与图片库一致的参考索引库，得到的结果却准确多了。 -- 这一点，可能仅是规律
	clipIndexToIdentSeeker := NewMultyClipBIndexToIdentSeeker([]*DBConfig {InitIndexToClipDB(dbId)})

	defer clipIndexToIdentSeeker.Close()

	//各个子图的 imgIndexContainer 容器: 每个容器表示某些 img index 与相应的子图有关系(即子图也出现在这些 img 中)
	occedImgIndex = imgCache.NewMyMap(true)
	cachedImgIndexToIdent = imgCache.NewMyMap(false)
	allBranchesIndex = make([] [][]byte, len(clipIndexes))

	for i, clipIndex := range clipIndexes{
		//与第 i 个子图相似的子图
		branchesIndexes := ImgIndex.ClipIndexBranch(clipIndex)
		allBranchesIndex[i] = branchesIndexes
		if hasDuplicateClipWithBefore(allBranchesIndex, i){
			continue
		}

		clipIdentsSet := clipIndexToIdentSeeker.SeekRegionForBranches(branchesIndexes)

		if 0 == len(clipIdentsSet){
			fmt.Println("no same clip: ", i)
		}

		for _,clipIdents := range clipIdentsSet{
			if 0 == len(clipIdents){
				continue
			}
			for _,clipIdent := range clipIdents{
				if 0 == len(clipIdent){
					continue
				}

				//当前子图出现在下面的 img 中. 为了唯一性表示，使用 img index 作为键去表示
				imgIndexDBId := clipIdent[0]
				imgIdent := clipIdent[0:5]

				//当前图跳过
				if bytes.Equal(curImgIdent, imgIdent){
					continue
				}

				imgIndex := InitImgToIndexDB(imgIndexDBId).ReadFor(imgIdent)
				if 0 == len(imgIndex){
					fmt.Println("error, can't get index for img: ",imgIndexDBId,"_", string(ImgIndex.ParseImgKeyToPlainTxt(imgIdent[1:])))
					return
				}
				//与当前大图是同一张图
				if bytes.Equal(curImgIndex, imgIndex){
					continue
				}
				cachedImgIndexToIdent.Put(imgIndex, imgIdent)
				occedImgIndex.Put(imgIndex, uint8(i))	//第 i 个子图出现在 imgIndex 所指示的大图中
			}
		}
	}
	return
}


/*
	以 dbId 库中的 imgKey 为对象，找出 imgKey 中哪些子图共同出现在其它大图中. GetInitedClipIndexToIdentDB() 作为查找库(待参考的 clipIndexToIdent 库).
	算法思路为: 对 imgKey 每个子图建立分支索引，在参考库中查找各个索引对应的 clipIdent, 由 clipIdent 可直接得到 imgIdent, 即是当前子图出现的大图，令为“母图”
	计算第 i 个子图的母图集合，设为 occImgSet[i]
	计算 occImgSet 各个元素的交集，same = occImgSet[i] ∩ occImgSet[j], 若集合 same 不为空则子图 i 和 j 共同出现在 same 集合中的大图里面
*/
func SearchCoordinateForClip(dbId uint8, imgKey []byte) (whichesGroupAndCount *imgCache.MyMap, allBranchesIndex [] [][]byte ) {
	imgName := strconv.Itoa(int(dbId)) + "-" + string(ImgIndex.ParseImgKeyToPlainTxt(imgKey))
	//计算哪些大图中联合出现了 imgKey 中的多个子图. 注意 imgKey 不包含在内
	occedImgIndex, _, allBranchesIndex := occInImgs(dbId, imgKey)

	if nil == occedImgIndex || 0 == occedImgIndex.KeyCount(){
		return
	}

	var resWhiches [][]uint8

	motherImgIndexes := occedImgIndex.KeySet()

	//occedImgIndex 每一个键值对 key, value 表示: key 指示的大图中出现了 value指示的子图
	//motherImgIndexes 一个元素即代表一个母图，imgKey 中有子图同时出现在这个母图中.
	//汇总结果要注意将重复的计次, 如 i,j 这两个子图同时出现在母图 A 也出现在 B 中，则 i,j 联合出现次数应该为 2
	//另外，由于 imgKey 在之前的计算中被忽略，它里面所有的子图的组合都出现在了 imgKey 中，所以各个 group 在计次时需要加上 1

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
			countExclusiceCurrentImg := interfaceCounts[0].(int)
			whichesGroupAndCount.Put(whiches, countExclusiceCurrentImg + 1)
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

/**
	考虑到 occInImgs 的计算方式, 设计算得到：
	A --> 1,3,4
	B --> 1,3,4
	C --> 3,4
	D --> 3,4
	举例说明: 3,4 出现在了 ABCD 四个母图中, 支持度应该为 4.
	所以我们应该得到最终的结果(以 whiches 为键)
	3,4 	--> 4(A,B,C,D)
	1,3,4	--> 2(A,B,)
 */
func statCoordinateResult(whichesGroups[] []byte) *imgCache.MyMap {
	resMap := imgCache.NewMyMap(false)

	//按长度逆序排列. 先处理较长的 whiches: 若后面较短的 whiches 被前面的包含则执行注释中的逻辑: 累加次数
	fileUtil.BytesArraySortByLengthDesc(whichesGroups)
	bytesMarker := newSlotMark(int(config.CLIP_COUNTS_OF_IMG))

	for _,whiches := range whichesGroups{
		if 0 == resMap.KeyCount() || !bytesMarker.isBytesMarked(whiches){
			resMap.Put(whiches, 1)
		}else{
			curGroupCount := 0
			exsitsKeys := resMap.KeySet()
			for _,key := range exsitsKeys{
				if -1 != bytes.Index(key, whiches){
					curGroupCount ++
				}
			}
			curGroupCount ++
			resMap.Put(whiches, curGroupCount)
		}

		bytesMarker.markBytesInSlotMap(whiches)
	}

	return resMap
}

type slotMark struct {
	slotmap []bool
}

func newSlotMark(maxSlot int) *slotMark {
	return &slotMark{slotmap:make([]bool, maxSlot)}
}

func (this *slotMark)markBytesInSlotMap(whiches []byte)  {
	for _,which := range whiches{
		this.slotmap[int(which)] = true
	}
}

func (this *slotMark) isBytesMarked(whiches []byte) bool {
	for _,which := range whiches{
		if false == this.slotmap[int(which)]{
			return false
		}
	}
	return true
}


func getPrintSearchResult(imgIndex []byte, cachedImgIndexToIdent *imgCache.MyMap, whiches []uint8 ) string {
	buff := ""
	buff += "{"
	interfaceOccedImgIdent := cachedImgIndexToIdent.Get(imgIndex)
	if 1 != len(interfaceOccedImgIdent){
		fmt.Println("oops, cache not contain imgIndex ------- ")
		return ""
	}
	occedImgIdent := interfaceOccedImgIdent[0].([]byte)
	occedImgName := strconv.Itoa(int(occedImgIdent[0])) + "-" + string(ImgIndex.ParseImgKeyToPlainTxt(occedImgIdent[1: ImgIndex.IMG_IDENT_LENGTH]))
	buff += occedImgName + "-"
	buff += "["
	for _,which := range whiches{
		buff += strconv.Itoa(int(which)) + ","
	}
	buff += "], "

	buff += "}, "
	return buff
}
