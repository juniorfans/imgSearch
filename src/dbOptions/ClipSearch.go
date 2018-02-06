package dbOptions

import (
	"imgIndex"
	"util"
	"github.com/syndtr/goleveldb/leveldb/util"
	"fmt"
	"config"
	"imgCache"
	"bytes"
)

//分析 dbId 库中的 imgKey 指示的图片: 其子图有没有共同出现在其它的图中
func SearchClipsOfImg(dbId uint8, imgKey []byte)  {
	targetImgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	targetImgIdent[0] = byte(dbId)
	copy(targetImgIdent[1:], imgKey)

	targetImgIndex := InitMuImgToIndexDb(dbId).ReadFor(targetImgIdent)
	if 0 == len(targetImgIndex){
		fmt.Println("can't find index for img: ", dbId , "_" ,string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)))
		return
	}


	clipConfig := config.GetClipConfigById(0)
	clipIndexes := GetDBIndexOfClips(PickImgDB(dbId) , imgKey, clipConfig.ClipOffsets, clipConfig.ClipLengh)
	clipSeeker := NewMultyDBIndexSeeker(GetInitedClipIndexToIdentDB())
	//imIdentToIndexDBs := NewMultyDBReader(GetInitedImgIdentToIndexDB())

	//哪些大图中出现了目标大图中的子图. 使用这些大图的 indexBytes 作为 key, 目标大图的子图之 ident 作为 value
	imgIndexContainsClips := make([]*imgCache.MyMap, len(clipIndexes))

	for i, clipIndex := range clipIndexes{

		//查询 clipIndex 的分支索引对应的所有 clipIdent
		//计算出它们都在的大图. 使用大图的 index 作为键, 小图的 index 作为值. 这样可以真实统计出一个小图出现在哪些大图中
		imgIdentContainsClip := imgCache.NewMyMap(false)	//此处使用单一值的 map 意义是若当前子图出现在了 index 相同的大图中，则只计算一次
		imgIndexContainsClips[i] = imgIdentContainsClip

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

				imgIndex := InitMuImgToIndexDb(imgIndexDBId).ReadFor(imgIdent)
				if 0 == len(imgIndex){
					fmt.Println("error, can't get index for img: ",imgIndexDBId,"_", string(ImgIndex.ParseImgKeyToPlainTxt(imgIdent[1:])))
					return
				}
				imgIdentContainsClip.Put(imgIndex, uint8(i))	//第 i 个子图出现在 imgIndex 所指示的大图中

			//	fmt.Println("------------------------",i," in ")
			//	fileUtil.PrintBytesLimit(imgIndex, 10)
			}
		}
	}

	//本目标大图中有哪些子图出现在了某一个相同的大图中. 待求大图的 indexBytes 作为键, 目标大图之子图的 ident 作为值
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

	keySet := whichInTheSameImg.KeySet()
	for _,key := range keySet{
		values := whichInTheSameImg.Get(key)
		if 0 == len(values){
			continue
		}

		if 1 < len(values) {

			//八个子图也同时出现在了另外的大图中, 需要判断此大图的 indexBytes 是否与 target 一致
			if 8 != len(values) || !bytes.Equal(key, targetImgIndex){

				fmt.Println("find clip in other same img: ")
				//whichs 共同出现在 key 所指示的大图中
				whichs := make([]uint8, len(values))
				for i,v := range values{
					whichs[i] = v.(uint8)
				}

				fmt.Println("list: ", whichs)
			}
		}
	}

}


func SearchClipEx(clipIndexBytes []byte)  {
	dbs := GetInitedClipIndexToIdentDB()
	seeker := NewMultyDBIndexSeeker(dbs)
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

		if len(curIndex) == ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_STAT_BYTES_LEN{
	//		fmt.Print("curIndex: ")
	//		fileUtil.PrintBytes(curIndex)
			for _, branchIndex := range branchesIndexes{
				if IsSameIndex(curIndex, branchIndex){
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
