package dbOptions

import (
	"imgCache"
	"imgIndex"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math"
	"fmt"
	"util"
)


//从多个表里面读取 key- value
type MultyDBClipBranchIndexToIdentSeeker struct {
	seekRes chan [][]byte
	dbs []*DBConfig
}

func NewMultyClipBIndexToIdentSeeker(dbs []*DBConfig) *MultyDBClipBranchIndexToIdentSeeker {

	if 0 == len(dbs){
		return nil
	}

	ret := MultyDBClipBranchIndexToIdentSeeker{}
	ret.dbs = dbs
	ret.seekRes = make(chan [][]byte, len(dbs))
	return &ret
}

func (this *MultyDBClipBranchIndexToIdentSeeker) Close ()  {
	close(this.seekRes)
}


//每一个 clipIndex 可能对应多个 clipIdent, 此函数用于返回一个 clipIndex 的分支索引对应的全部 clipIdent: n 个 [][]byte, 每一个是某些 clipIdent
func (this *MultyDBClipBranchIndexToIdentSeeker) SeekRegionForBranches (branchesIndexes [][]byte) [][][]byte {

	//填坑: branchesIndexes 的长度有可能是 1: 当 clipIndexBytes 的分支字节值已经是最大时(>=250).
	//此时要注意 minBranch 和 maxBranch 的取值
	branchBitsArray := make([][]byte, len(branchesIndexes))
	minBranch := []byte{255,255}
	maxBranch := []byte{0,0}

	for i, branchIndex := range branchesIndexes{
		//	fmt.Print("branchIndex: ")
		//	fileUtil.PrintBytes(branchIndex)
		branchBitsArray[i] = fileUtil.CopyBytesTo(branchIndex[ : ImgIndex.CLIP_INDEX_BRANCH_BITS])
	}

	if 1 < len(branchBitsArray){
		for _,b:=range branchBitsArray {
			if fileUtil.BytesCompare(minBranch, b) > 0 {
				minBranch = b
			}
			if fileUtil.BytesCompare(maxBranch, b) < 0 {
				maxBranch = b
			}
		}
	}else{
		minBranch = branchBitsArray[0]
		maxBranch = []byte{255,255}	//在分支字节索引中, 最大的值是 250
	}


	region := util.Range{Start:minBranch, Limit:maxBranch}

	for i:=0;i < len(this.dbs);i ++{
		go this.seekRegion(this.dbs[i], &region, branchesIndexes)
	}

	ret := make([][][]byte, len(this.dbs))
	for i:=0; i < len(this.dbs); i++{
		r := <- this.seekRes
		ret[i] = r
	}

	return ret

}

//每一个 clipIndex 可能对应多个 clipIdent, 此函数用于返回一个 clipIndex 的分支索引对应的全部 clipIdent: n 个 [][]byte, 每一个是某些 clipIdent
func (this *MultyDBClipBranchIndexToIdentSeeker) SeekRegion (clipIndexBytes []byte) [][][]byte {
	branchesIndexes := ImgIndex.ClipIndexBranch(clipIndexBytes)
	return this.SeekRegionForBranches(branchesIndexes)
}

func (this *MultyDBClipBranchIndexToIdentSeeker) seekRegion(db *DBConfig, region *util.Range, branchesIndexes [][]byte)  {
	res := imgCache.NewMyMap(false)

	iter := db.DBPtr.NewIterator(region, &db.ReadOptions)
	iter.First()
	for iter.Valid(){
		curIndex := iter.Key()

		if len(curIndex) == ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN{
			for _, branchIndex := range branchesIndexes{
				if IsSameClipBranchIndex(curIndex, branchIndex){
					clipIdents := fileUtil.CopyBytesTo(iter.Value())
					clen := len(clipIdents)
					for i:=0;i < clen;i += ImgIndex.IMG_CLIP_IDENT_LENGTH{
						res.Put(clipIdents[i:i+ImgIndex.IMG_CLIP_IDENT_LENGTH], 1)
					}
				}
			}
		}

		iter.Next()
	}

	iter.Release()

	this.seekRes <- res.KeySet()
}


//-------------------------------------------------------------------------------------------

func IsSameClipBranchIndex(leftIndex, rightIndex[]byte) bool {
	if len(leftIndex) != len(rightIndex){
		fmt.Println("error, left, right len not equal as the clip index")
		return false
	}
	if len(leftIndex) != ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN{
		return false
	}

	leftSD := leftIndex[2]
	rightSD := rightIndex[2]

	leftMean := leftIndex[3]
	rightMean := rightIndex[3]

	searchConf := ImgIndex.TheclipSearchConf

	meanDiff := math.Abs(float64(leftMean-rightMean))
	if searchConf.Delta_mean < meanDiff{
		return false
	}

	sdDiff := math.Abs(float64(leftSD - rightSD))
	if searchConf.Delta_sd < sdDiff{
		return false
	}

	return isSameClip(leftIndex, rightIndex)
}