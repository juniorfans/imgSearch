package dbOptions

import (
	"imgIndex"
	"util"
	"github.com/syndtr/goleveldb/leveldb/util"
	"fmt"
)

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

			fmt.Println(dbId, string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)), which)
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
