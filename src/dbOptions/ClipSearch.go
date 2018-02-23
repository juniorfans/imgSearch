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

//第 cur 个 clip branches 是否与前面的重复
func isDuplicateClipBefore(allBranches [][][]byte, cur int) bool {
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
	查找 dbId 中的 imgKey 中的各个子图与哪些大图中的子图的重叠
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
		if isDuplicateClipBefore(allBranchesIndex, i){
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

				//当前子图(从 target img 计算而来)必然已经在 target img 中, 跳过
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

func SearchEx(dbId uint8, imgKey []byte) (resWhiches []uint8, allBranchesIndex [] [][]byte ) {
	buff := ""
	imgName := strconv.Itoa(int(dbId)) + "-" + string(ImgIndex.ParseImgKeyToPlainTxt(imgKey))
	buff += imgName + ": "

	occedImgIndex, cachedImgIndexToIdent, allBranchesIndex := occInImgs(dbId, imgKey)

	imgIndexes := occedImgIndex.KeySet()
	show := false
	for _,imgIndex := range imgIndexes{
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
		resWhiches = whiches

		buff += "{"
		interfaceOccedImgIdent := cachedImgIndexToIdent.Get(imgIndex)
		if 1 != len(interfaceOccedImgIdent){
			fmt.Println("oops, bug ------- ", imgName)
		}
		occedImgIdent := interfaceOccedImgIdent[0].([]byte)
		occedImgName := strconv.Itoa(int(occedImgIdent[0])) + "-" + string(ImgIndex.ParseImgKeyToPlainTxt(occedImgIdent[1: ImgIndex.IMG_IDENT_LENGTH]))
		buff += occedImgName + "-"
		buff += "["
		for _,which := range whiches{
			buff += strconv.Itoa(int(which)) + ","
		}
		buff += "]"

		buff += "}, "
		show = true
	}
	if show{
		fmt.Println(buff)
	}

	return
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
