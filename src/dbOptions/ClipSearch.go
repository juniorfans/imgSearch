package dbOptions

import (
	"imgIndex"
	"util"
	"github.com/syndtr/goleveldb/leveldb/util"
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
 */
func occInImgs(dbId uint8, imgKey []byte) (occedImgIndex *imgCache.MyMap, cachedImgIndexToIdent *imgCache.MyMap, allBranchesIndex[] [][]byte ){
	curImgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	curImgIdent[0] = byte(dbId)
	copy(curImgIdent[1:], imgKey)
	curImgIndex := InitMuImgToIndexDB(uint8(dbId)).ReadFor(curImgIdent)

	clipConfig := config.GetClipConfigById(0)
	clipIndexes := GetDBIndexOfClips(PickImgDB(dbId) , imgKey, clipConfig.ClipOffsets, clipConfig.ClipLengh)
	clipSeeker := NewMultyClipBIndexToIdentSeeker(GetInitedClipIndexToIdentDB())
	//各个子图的 imgIndexContainer 容器: 每个容器表示某些 img index 与相应的子图有关系(即子图也出现在这些 img 中)
	occedImgIndex = imgCache.NewMyMap(true)
	cachedImgIndexToIdent = imgCache.NewMyMap(false)
	allBranchesIndex = make([] [][]byte, len(clipIndexes))

	for i, clipIndex := range clipIndexes{
		//与第 i 个子图相似的子图
		branchesIndexes := ImgIndex.ClipIndexBranch(clipIndex.GetIndexBytesIn3Chanel())
		allBranchesIndex[i] = branchesIndexes
		if hasDuplicateClipWithBefore(allBranchesIndex, i){
			continue
		}

		clipIdentsSet := clipSeeker.SeekRegionForBranches(branchesIndexes)

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

				imgIndex := InitMuImgToIndexDB(imgIndexDBId).ReadFor(imgIdent)
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
//	buff := ""
	imgName := strconv.Itoa(int(dbId)) + "-" + string(ImgIndex.ParseImgKeyToPlainTxt(imgKey))
//	buff += imgName + ": "

	//计算哪些大图中联合出现了 imgKey 中的多个子图. 注意 imgKey 不包含在内
	occedImgIndex, _, allBranchesIndex := occInImgs(dbId, imgKey)

	if 0 == occedImgIndex.KeyCount(){
		return
	}

	resWhiches := make([][]uint8, occedImgIndex.KeyCount())
	groupCount := 0

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

		resWhiches[groupCount] = whiches
		groupCount ++

	//	buff += getPrintSearchResult(imgIndex, cachedImgIndexToIdent, whiches)

	}

	//校准次数. 注意校验不能在 statCoordinateResult 中边统计边校准：只能最终校准.
	//原因在于: 设 1,3,4 同时出现在 A, B 图中，3,4 同时出现在 A,B,C 中则
	resWhiches = resWhiches[ : groupCount]
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
	for _,whiches := range whichesGroups{
		if 0 == resMap.KeyCount(){
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
	}



	return resMap
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

//分析 dbId 库中的 imgKey 指示的图片: 其子图有没有共同出现在其它的图中
func SearchClipsOfImg(dbId uint8, imgKey []byte) (resWhiches []uint8, clipIndexes []ImgIndex.SubImgIndex ) {
	targetImgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	targetImgIdent[0] = byte(dbId)
	copy(targetImgIdent[1:], imgKey)

	targetImgIndex := InitMuImgToIndexDB(dbId).ReadFor(targetImgIdent)
	if 0 == len(targetImgIndex){
		fmt.Println("can't find index for img: ", dbId , "_" ,string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)))
		return
	}


	clipConfig := config.GetClipConfigById(0)
	clipIndexes = GetDBIndexOfClips(PickImgDB(dbId) , imgKey, clipConfig.ClipOffsets, clipConfig.ClipLengh)
	clipSeeker := NewMultyClipBIndexToIdentSeeker(GetInitedClipIndexToIdentDB())
	//imIdentToIndexDBs := NewMultyDBReader(GetInitedImgIdentToIndexDB())

	//各个子图的 imgIndexContainer 容器: 每个容器表示某些 img index 与相应的子图有关系(即子图也出现在这些 img 中)
	imgIndexContainsClips := make([]*imgCache.MyMap, len(clipIndexes))

	indexToImgIdentReader := NewMultyDBReader(GetInitedImgIndexToIdentDB())

	for i, clipIndex := range clipIndexes{
		//计算出大图的各个小图的分支索引，各出现在哪些大图中

		//imgIndex --> which.
		curImgIndexContainer := imgCache.NewMyMap(false)
		imgIndexContainsClips[i] = curImgIndexContainer

		//与第 i 个子图相似的子图
		clipIdentsSet := clipSeeker.SeekRegion(clipIndex.GetIndexBytesIn3Chanel())
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

				//当前子图(从 target img 计算而来)必然已经在 target img 中, 跳过
				if bytes.Equal(targetImgIdent, imgIdent){
					continue
				}

				imgIndex := InitMuImgToIndexDB(imgIndexDBId).ReadFor(imgIdent)
				if 0 == len(imgIndex){
					fmt.Println("error, can't get index for img: ",imgIndexDBId,"_", string(ImgIndex.ParseImgKeyToPlainTxt(imgIdent[1:])))
					return
				}
				curImgIndexContainer.Put(imgIndex, uint8(i))	//第 i 个子图出现在 imgIndex 所指示的大图中
			}
		}
	}

	//汇总相同大图中出现了哪些子图. 键: imgIndex, 值: which
	whichInTheSameImg := imgCache.NewMyMap(true)	//此处使用多值 map 的意义是需要计算相同 index 的大图包括了哪些子图.

	for _, imgMap := range imgIndexContainsClips {
		keySet := imgMap.KeySet()
		if 0 == len(keySet){
			continue
		}
		for _,key := range keySet{
			value := imgMap.Get(key)
			if 0 == len(value){
				continue
			}
			whichInTheSameImg.Put(key, value[0])
		}

		imgMap.Destroy()
	}

	//汇总某些子图出现在相同的大图中
	clipInImg := imgCache.NewMyMap(true)

	keySet := whichInTheSameImg.KeySet()
	for _,key := range keySet {
		values := whichInTheSameImg.Get(key)
		if 2 > len(values) {
			continue
		}

		clipKey := make([]byte, len(values))
		for i,v := range values{
			clipKey[i] = (v.(uint8))
		}

		fileUtil.BytesSort(clipKey)

		//若 0, 4, 6 这三个子图出现在 key 表示的大图中, 为了准备计算 0,4 联合出现的次数， 4,6 联合出现的次数， 0,4,6 联合出现的次数
		//需要进行分别计算
		for i:=2;i <= len(clipKey);i ++{
			clipInImg.Put(clipKey[:i], key)	//子图 [0, i) 出现在 key 代表的大图中
		}
	}

	imgName := strconv.Itoa(int(dbId)) + "_" + string(ImgIndex.ParseImgKeyToPlainTxt(imgKey))
	fmt.Print("find clip in other same img: ", imgName, " ---- ")

	keySet = clipInImg.KeySet()
	for _,key := range keySet{
		values := clipInImg.Get(key)
		if 1 > len(values){
			continue
		}

		//判断同构图. 八个子图也同时出现在了另外的大图中, 需要判断此大图的 indexBytes 是否与 target 一致
		if 8 != len(key){
			fmt.Print(key, " : ", len(values))	//len(values) 是出现的次数

			for _,v := range values{
				imgIndex := v.([]byte)
				imgIdents := indexToImgIdentReader.ReadFor(imgIndex)

				if len(imgIdents) > 0 && 0 == len(imgIdents[0]) % ImgIndex.IMG_IDENT_LENGTH{	//img index to ident 库的 value 可能有 n 个 imgident
					fmt.Print(" | ", strconv.Itoa(int(imgIdents[0][0])),"_",string(ImgIndex.ParseImgKeyToPlainTxt(imgIdents[0][1:ImgIndex.IMG_IDENT_LENGTH])))
				}
			}
			fmt.Println()
			//todo 此处不正确，resWhiches 会被赋值多次, 如当 0,4,6 联合出现的次数与 4,6 联合出现的次数不一致时
			resWhiches = key
		}
	}
	return
}


func SearchClipEx(clipIndexBytes []byte)  {
	dbs := GetInitedClipIndexToIdentDB()
	seeker := NewMultyClipBIndexToIdentSeeker(dbs)
	resSet := seeker.SeekRegion(clipIndexBytes)
	for _,res := range resSet{
		if 0 == len(res){
			continue
		}
		for _,r := range res{
			if 0 == len(r){
				continue
			}

			dbId := uint8(r[0])
			imgKey := r[1:5]
			which := uint8(r[5])
			indexes := GetDBIndexOfClips(PickImgDB(dbId) , imgKey, []int{-1} ,-1)
			SaveClipsAsJpg("E:/gen/search/", indexes[which])
			SaveAImg(dbId, imgKey, "E:/gen/search/img/")

			//	fmt.Println(dbId, string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)), which)
		}
	}
}

//我们认为 index 值是大端模式
func SearchClip(clipIndexBytes []byte) {

	indexToClipDB := InitMuIndexToClipDB(2)


	branchesIndexes := ImgIndex.ClipIndexBranch(clipIndexBytes)

	branchBitsArray := make([][]byte, len(branchesIndexes))
	minBranch := []byte{255,255}
	maxBranch := []byte{0,0}
	for i, branchIndex := range branchesIndexes{
	//	fmt.Print("branchIndex: ")
	//	fileUtil.PrintBytes(branchIndex)
		branchBitsArray[i] = fileUtil.CopyBytesTo(branchIndex[ : ImgIndex.CLIP_INDEX_BRANCH_BITS])
	}

	for _,b:=range branchBitsArray {
		if fileUtil.BytesCompare(minBranch, b) > 0 {
			minBranch = b
		}
		if fileUtil.BytesCompare(maxBranch, b) < 0 {
			maxBranch = b
		}
	}

	sameClip := make(map[string]int)

	limit := fileUtil.CopyBytesTo(maxBranch)
	fileUtil.BytesIncrement(limit)
	ir := util.Range{Start:minBranch, Limit: limit}
	iter := indexToClipDB.DBPtr.NewIterator(&ir,&indexToClipDB.ReadOptions)
	iter.First()
	for iter.Valid(){
		curIndex := iter.Key()

		if len(curIndex) == ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN{
	//		fmt.Print("curIndex: ")
	//		fileUtil.PrintBytes(curIndex)
			for _, branchIndex := range branchesIndexes{
				if IsSameClipBranchIndex(curIndex, branchIndex){
				//	fmt.Println("find same: ----------------------------------")
				//	fileUtil.PrintBytes(curIndex)
				//	fileUtil.PrintBytes(branchIndex)

					clips := ImgIndex.FromClipIdentsToStrings(iter.Value())
					for _, clip := range clips{
						sameClip[clip] ++
					}

				}
			}
		}

		iter.Next()
	}

	for clip, _ := range sameClip{
		fmt.Println(clip)
	}
}
