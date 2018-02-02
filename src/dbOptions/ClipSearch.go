package dbOptions

import (
	"imgIndex"
	"util"
	"github.com/syndtr/goleveldb/leveldb/util"
	"fmt"
	"math"
)

//我们认为 index 值是大端模式
func SearchClip(clipIndexBytes []byte) {

	indexToClipDB := InitMuIndexToClipDB(2)//GetTotalMuIndexToClipDB()


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

var Delta_sd = 5.0
var Delta_mean = 5.0
var Delta_Eul = 10.0

func IsSameIndex(leftIndex, rightIndex[]byte) bool {
	if len(leftIndex) != len(rightIndex){
		fmt.Println("error, left, right len not equal as the clip index")
		return false
	}
	if len(leftIndex) != ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_STAT_BYTES_LEN{
		return false
	}

	leftSD := leftIndex[2]
	rightSD := rightIndex[2]

	leftMean := leftIndex[3]
	rightMean := rightIndex[3]

	if Delta_mean < math.Abs(float64(leftMean-rightMean)){
		return false
	}
	if Delta_sd < math.Abs(float64(leftSD - rightSD)){
		return false
	}

	return true

	//欧式距离
	sim := float64(0)
	for i:=0;i < len(leftIndex);i++{
		sim += math.Pow(float64(leftIndex[i]-rightIndex[i]), 2)
	}

	diff := math.Pow(sim / float64(len(leftIndex)), 0.5)
	return diff < Delta_Eul
}
