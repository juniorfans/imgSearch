package dbOptions

import (
	"imgCache"
	"imgIndex"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math"
	"fmt"
	"util"
	"config"
)


//从多个表里面读取 key- value
type MultyDBIndexSeeker struct {
	seekRes chan [][]byte
	dbs []*DBConfig
}

func NewMultyDBIndexSeeker(dbs []*DBConfig) *MultyDBIndexSeeker {

	if 0 == len(dbs){
		return nil
	}

	ret := MultyDBIndexSeeker{}
	ret.dbs = dbs
	ret.seekRes = make(chan [][]byte, len(dbs))
	return &ret
}

func (this *MultyDBIndexSeeker) Close ()  {
	close(this.seekRes)
}

//每一个 clipIndex 可能对应多个 clipIdent, 此函数用于返回一个 clipIndex 的分支索引对应的全部 clipIdent: n 个 [][]byte, 每一个是某些 clipIdent
func (this *MultyDBIndexSeeker) SeekRegion (clipIndexBytes []byte) [][][]byte {
	branchesIndexes := ImgIndex.ClipIndexBranch(clipIndexBytes)

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

func (this *MultyDBIndexSeeker) seekRegion(db *DBConfig, region *util.Range, branchesIndexes [][]byte)  {
	res := imgCache.NewMyMap(false)

	iter := db.DBPtr.NewIterator(region, &db.ReadOptions)
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

	this.seekRes <- res.KeySet()
}


//-------------------------------------------------------------------------------------------

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

	searchConf := config.ReadClipSearchConf("clip_search_conf.txt")

	meanDiff := math.Abs(float64(leftMean-rightMean))
	if searchConf.Delta_mean < meanDiff{
		return false
	}

	sdDiff := math.Abs(float64(leftSD - rightSD))
	if searchConf.Delta_sd < sdDiff{
		return false
	}

	//欧式距离
	sim := float64(0)
	for i:=0;i < len(leftIndex);i++{
		sim += math.Pow(float64(leftIndex[i]-rightIndex[i]), 2)
	}

	eulDiff := math.Pow(sim / float64(len(leftIndex)), 0.5)

	if searchConf.Delta_Eul < eulDiff{
		return false
	}

//	fmt.Println("meanDiff: ", meanDiff,", sdDiff: ", sdDiff, ", eulDiff: ", eulDiff)

	return true
}