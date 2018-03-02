package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"config"
	"strconv"
	"util"
	"github.com/pkg/errors"
	"imgCache"
	"imgIndex"
	"bufio"
	"os"
	"strings"
	"bytes"
	"sort"
)

var CLIP_VIRTUAL_TAGID_LEN = 10
var MAX_CLIP_VIRTUAL_TAGID = []byte{255,255,255,255,255,255,255,255,255,255}


//-------------------------------------------------------------------------------------------------
func ClipCoordinateLastDealedFor(threadId uint8) []byte {
	lastDealedKeyName := string(config.STAT_KEY_PREX) + "_LAST_DEALED_IMGKEY_" + strconv.Itoa(int(threadId))
	res := InitCoordinateClipToVTagMiddleDB().ReadFor([]byte(lastDealedKeyName))
	if len(res) == 0{
		return nil
	}
	return res
}

func SetClipCoordinateLastDealedFor(lastDealed []byte, threadId uint8)  {
	lastDealedKeyName := string(config.STAT_KEY_PREX) + "_LAST_DEALED_IMGKEY_" + strconv.Itoa(int(threadId))
	InitCoordinateClipToVTagMiddleDB().WriteTo([]byte(lastDealedKeyName), lastDealed)
}

func GetUnusedVirtualTagId(threadId uint8) (ret []byte , err error){
	lastUsedVirtualTagIdName := string(config.STAT_KEY_PREX) + "_UNUSED_VIRTUAL_TAG_ID_" + strconv.Itoa(int(threadId))
	res := InitCoordinateClipToVTagMiddleDB().ReadFor([]byte(lastUsedVirtualTagIdName))
	if len(res) == 0{
		//此处与 CLIP_VIRTUAL_TAGID_LEN 相对应
		return []byte{threadId, '_', 0,0,0,0,0,0,0,0}, nil
	}
	ret = fileUtil.CopyBytesTo(res)
	if fileUtil.BytesIncrement(ret){
		return ret, nil
	}else{
		return nil, errors.New("increment last used virtual tag error")
	}
}

func SetUnusedVirtualTagId(lastUsed []byte, threadId uint8)  {
	lastUsedVirtualTagIdName := string(config.STAT_KEY_PREX) + "_UNUSED_VIRTUAL_TAG_ID_" + strconv.Itoa(int(threadId))
	InitCoordinateClipToVTagMiddleDB().WriteTo([]byte(lastUsedVirtualTagIdName), lastUsed)
}


func CalCoordinateForDB(dbId uint8, dealCounts int)  {
	//初始化线程缓存及总体缓存. 线程缓存满了之后即加入总体缓存，则总体缓存统一写入内存
	coordinateCacheList := imgCache.KeyValueCacheList{}
	var callBack imgCache.CacheFlushCallBack = &ClipCoordinateCacheFlushCallBack{}
	coordinateCacheList.Init(true, &callBack,true, 100)

	threadToVirtualTagIds := make(map[uint8] []byte)
	for i:=0;i < config.MAX_THREAD_COUNT;i ++{
		vTagId, err := GetUnusedVirtualTagId(uint8(i))
		if nil == err{
			threadToVirtualTagIds[uint8(i)] = vTagId
		}
	}

	var visitCallBack VisitCallBack = &ClipCoordinateVisitCallBack{maxVisitCount:dealCounts,
		visitParams:ClipCoordinateVisitParams{dbId:dbId,
			cacheList: coordinateCacheList, threadVirtualTagIds: threadToVirtualTagIds}}

	VisitBySeek(PickImgDB(dbId), visitCallBack, -1)

	coordinateCacheList.FlushRemainKVCaches()
}

//-------------------------------------------------------------------------------------
type ClipCoordinateResult struct {
	clipIndexAndIdents   [][]byte

	whichesGroupAndCount *imgCache.MyMap

	allStatIndex         [] [][]byte

	beginVituralTagId    []byte
}

func (this *ClipCoordinateResult) PutToCacheList(threadId int, cacheList *imgCache.KeyValueCacheList) {
	var vTagId = this.beginVituralTagId

	KeyLen := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH) + CLIP_VIRTUAL_TAGID_LEN + 4
	if nil != this.whichesGroupAndCount && this.whichesGroupAndCount.KeyCount() != 0{
		//每个组中的 whiches 即是同时出现在某些大图中的


		theWhichesGroups := this.whichesGroupAndCount.KeySet()
		for _,whiches := range theWhichesGroups{
			interfaceCounts := this.whichesGroupAndCount.Get(whiches)
			if 1 != len(interfaceCounts){
				continue
			}

			//当前 group 支持度.
			curGroupSupport := interfaceCounts[0].(int)

			//whiches 是一个组, 有一个共同的 tag. 将它们两两配对: C2
			combineWhichList := fileUtil.Combine2(whiches)
			for _,combineWhiches := range combineWhichList{
				left := combineWhiches[0]
				right := combineWhiches[1]
				leftBranches := this.allStatIndex[int(left)]
				rightBranches := this.allStatIndex[int(right)]

				leftSourceIndexAndIdent := this.clipIndexAndIdents[int(left)]
				rightSourceIndexAndIdent := this.clipIndexAndIdents[int(right)]

				for _,lb := range leftBranches{
					for _, rb := range rightBranches{
						resKey := make([]byte, KeyLen)
						ci := 0
						ci += copy(resKey[ci:], lb)
						ci += copy(resKey[ci:], rb)
						ci += copy(resKey[ci:], leftSourceIndexAndIdent)
						ci += copy(resKey[ci:], rightSourceIndexAndIdent)
						ci += copy(resKey[ci:], vTagId)
						ci += copy(resKey[ci:], ImgIndex.Int32ToBytes(curGroupSupport))
						cacheList.Add(threadId, resKey, nil)
					}
				}
			}
			fileUtil.BytesIncrement(vTagId)
		}
	}
}

//--------------------------------------------------------------------------------------

type ClipCoordinateCacheFlushCallBack struct {

}


/**
	写入协同计算结果
	branchIndex1 | branchIndex2 | index+ident1 | index+ident2 | vtag --> support
	vtag | index+ident1 | index+ident2 --> support
 */
func (this *ClipCoordinateCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {
	keys := kvCache.KeySet()
	if len(keys) == 0{
		return true
	}
	biToTagDB := InitCoordinateClipToVTagMiddleDB() //InitClipCoordinateIndexToVTagIdDB()
	tagToBiDB := InitCoordinatevTagToClipDB()

	branchIndexToVTag := leveldb.Batch{}
	vTagToBranchIndex := leveldb.Batch{}

	keyLen := 2*ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + 2*(ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH) + CLIP_VIRTUAL_TAGID_LEN + 4

	anotherBiToVtagBuff := make([]byte, keyLen)
	vTagToBiBuff := make([]byte,  CLIP_VIRTUAL_TAGID_LEN+2*(ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH))
	//key 是 stat branch index1 | stat branch index2 | clip index+ident 1 | clip index+ident 2 | virtual tag ids , value 是 support
	for _,key := range keys{
		if len(key) != keyLen{
			fmt.Println("error, in coordinate flush cache, key length is not equal to expected, ", len(key), "!=", keyLen)
			continue
		}

		supportBytes := key[keyLen-4:]	//最后四个字节是 supportBytes

		anotherFormBiToVtag(key, anotherBiToVtagBuff)

		branchIndexToVTag.Put(anotherBiToVtagBuff, nil)
		branchIndexToVTag.Put(key, nil)


		transformBiVtagToVtagBi(key, vTagToBiBuff)
		vTagToBranchIndex.Put(vTagToBiBuff, supportBytes)
	}

	fmt.Println("write to branchIndexToVTag: ", branchIndexToVTag.Len())
	biToTagDB.WriteBatchTo(&branchIndexToVTag)
	fmt.Println("write to vTagToBranchIndex: ", vTagToBranchIndex.Len())
	tagToBiDB.WriteBatchTo(&vTagToBranchIndex)

	return true
}


//bitoTagBytes 是 stat branch index1 | stat branch index2 | clip index+ident 1 | clip index+ident 2 | virtual tag id , value 是 support
//tagToBiBytes 是 virtual tag id | clip index 1| clip index 2
func transformBiVtagToVtagBi(bitoTagBytes []byte, buff []byte) {
	vtagStart := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH)
	vtagLimit := vtagStart + CLIP_VIRTUAL_TAGID_LEN
	vtagId := bitoTagBytes[vtagStart : vtagLimit]

	clipIndexAndIdent1Start := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
	clipIndexAndIdent2Start := clipIndexAndIdent1Start + (ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIndexAndIdent2Limit := clipIndexAndIdent2Start + (ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIndexAndIdent1 := bitoTagBytes[clipIndexAndIdent1Start : clipIndexAndIdent2Start]
	clipIndexAndIdent2 := bitoTagBytes[clipIndexAndIdent2Start :clipIndexAndIdent2Limit]

	ci := 0
	ci += copy(buff[ci:], vtagId)
	ci += copy(buff[ci:], clipIndexAndIdent1)
	ci += copy(buff[ci:], clipIndexAndIdent2)
}

//颠倒 1 和 2 的顺序
func anotherFormBiToVtag(bitoTagBytes []byte, buff []byte)  {
	statIndex1Start := 0
	statIndex2Start := ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
	statIndex2Limit := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN

	statIndex1 := bitoTagBytes[statIndex1Start: statIndex2Start]
	statIndex2 := bitoTagBytes[statIndex2Start: statIndex2Limit]

	clipIndexAndIdent1Start := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
	clipIndexAndIdent2Start := clipIndexAndIdent1Start + (ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIndexAndIdent2Limit := clipIndexAndIdent2Start + (ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIndexAndIdent1 := bitoTagBytes[clipIndexAndIdent1Start : clipIndexAndIdent2Start]
	clipIndexAndIdent2 := bitoTagBytes[clipIndexAndIdent2Start :clipIndexAndIdent2Limit]

	vtagStart := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH)

	copy(buff[statIndex1Start:], statIndex2)
	copy(buff[statIndex2Start:], statIndex1)
	copy(buff[clipIndexAndIdent1Start:], clipIndexAndIdent2)
	copy(buff[clipIndexAndIdent2Start:], clipIndexAndIdent1)

	//将 vtag 和 support 一并拷贝
	copy(buff[vtagStart:], bitoTagBytes[vtagStart:])
}

//-------------------------------------------------------------------------------------
type ClipCoordinateVisitParams struct {
	cacheList imgCache.KeyValueCacheList
	dbId uint8
	threadVirtualTagIds map[uint8][]byte	//各线程可以直接使用的 virtual tag id
}

type ClipCoordinateVisitCallBack struct {
	maxVisitCount int
	visitParams ClipCoordinateVisitParams
}

func (this* ClipCoordinateVisitCallBack)GetMaxVisitCount() int{
	return this.maxVisitCount
}


func (this* ClipCoordinateVisitCallBack) Visit(visitInfo *VisitIngInfo) bool{

	if 0 != visitInfo.curCount && visitInfo.curCount % 100 == 0{
		fmt.Println("thread ", visitInfo.threadId, " dealing ", visitInfo.curCount)
	}

	imgKey := visitInfo.key

	imgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	imgIdent[0] = this.visitParams.dbId
	copy(imgIdent[1:], imgKey)

	vTagId := this.visitParams.threadVirtualTagIds[uint8(visitInfo.threadId)]

	whichesGroupAndCount, clipIndexAndIdents, allstatBranchesIndex := SearchCoordinateForClipEx(this.visitParams.dbId, imgKey)

	res := &ClipCoordinateResult{
		whichesGroupAndCount:whichesGroupAndCount,
		clipIndexAndIdents:clipIndexAndIdents,
		allStatIndex:allstatBranchesIndex,
		beginVituralTagId:vTagId,
	}

	res.PutToCacheList(visitInfo.threadId, &this.visitParams.cacheList)
	return true
}

//遍历完成回调函数
func (this* ClipCoordinateVisitCallBack) VisitFinish(finishInfo * VisitFinishedInfo){
	SetClipCoordinateLastDealedFor(finishInfo.lastSuccessDealedKey, uint8(finishInfo.threadId))
	SetUnusedVirtualTagId(this.visitParams.threadVirtualTagIds[uint8(finishInfo.threadId)], uint8(finishInfo.threadId))

	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ImgIndex.ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}

func (this* ClipCoordinateVisitCallBack) GetLastVisitPos(dbId uint8, threadId int) []byte{
	return ClipCoordinateLastDealedFor(uint8(threadId))
}

func VerifyCoordinateResult()  {
	var dbIdsStr string
	var offset, limit int
	stdin := bufio.NewReader(os.Stdin)

	fmt.Print("input dbIds, split by dot: ")
	fmt.Fscan(stdin, &dbIdsStr)

	dbIdList := strings.Split(dbIdsStr, ",")
	dbIds := make([]uint8, len(dbIdList))
	for i,dbId := range dbIdList{
		curDBId ,_ := strconv.Atoi(dbId)
		if curDBId > 255{
			fmt.Println("dbid can't more than 255")
			return
		}
		dbIds[i] = uint8(curDBId)
	}

	fmt.Print("input offset and limit: ")
	fmt.Fscan(stdin, &offset, &limit)

	innerVerifyCoordinateResult(dbIds, offset, limit)
}

func testQueryAnyOneClipIdentByIndexAndIdent(dbId uint8, sourceIndexAndIdent []byte ) []byte {
	clipIndex := sourceIndexAndIdent[:ImgIndex.CLIP_INDEX_BYTES_LEN]
	theClipIdent := sourceIndexAndIdent[ImgIndex.CLIP_INDEX_BYTES_LEN:]

	statBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndex)
	var clipIdentList []byte
	var clipIdents [][]byte
	var curIndex []byte
	fnd := false

	for _, statIndex := range statBranchIndexes{
		clipIdentList = InitStatIndexToClipDB(dbId).ReadFor(statIndex)

		clipIdents = make([][]byte, len(clipIdentList) / ImgIndex.IMG_CLIP_IDENT_LENGTH)
		ci := 0
		for i:=0;i < len(clipIdentList); i += ImgIndex.IMG_CLIP_IDENT_LENGTH{
			clipIdents[ci] = fileUtil.CopyBytesTo(clipIdentList[i: i + ImgIndex.IMG_CLIP_IDENT_LENGTH])
			ci ++
		}

		for _,clipIdent := range clipIdents{
			if len(clipIdent) != ImgIndex.IMG_CLIP_IDENT_LENGTH{
				continue
			}

			if bytes.Equal(theClipIdent, clipIdent){
				curIndex = InitClipToIndexDB(clipIdent[0]).ReadFor(clipIdent)

				if bytes.Equal(curIndex, clipIndex){

					if isSameClip(curIndex, clipIndex){
						return clipIdent
					}else{
						fmt.Println("error logic, find it, but can't same")
					}

					fnd = true
				}
			}
		}
	}
	if fnd{

	}

	return nil
}

func QueryAnyOneClipIdentByIndex(dbId uint8,sourceIndex []byte) []byte {
	statBranchIndexes := ImgIndex.ClipStatIndexBranch(sourceIndex)
	var clipIdentList []byte
	var clipIdents [][]byte
	var curIndex []byte
	for _, statIndex := range statBranchIndexes{
		clipIdentList = InitStatIndexToClipDB(dbId).ReadFor(statIndex)

		clipIdents = make([][]byte, len(clipIdentList) / ImgIndex.IMG_CLIP_IDENT_LENGTH)
		ci := 0
		for i:=0;i < len(clipIdentList); i += ImgIndex.IMG_CLIP_IDENT_LENGTH{
			clipIdents[ci] = fileUtil.CopyBytesTo(clipIdentList[i: i + ImgIndex.IMG_CLIP_IDENT_LENGTH])
			ci ++
		}

		for _,clipIdent := range clipIdents{
			if len(clipIdent) != ImgIndex.IMG_CLIP_IDENT_LENGTH{
				continue
			}
			curIndex = InitClipToIndexDB(clipIdent[0]).ReadFor(clipIdent)

			if isSameClip(sourceIndex, curIndex){
				return clipIdent
			}
		}
	}
	return nil
}

//----------------------------------------------------------------------------
func innerVerifyCoordinateResult(indexBbIdReferenced []uint8, offset, limit int)  {

	if len(indexBbIdReferenced) == 0{
		return
	}
	for _,dbId := range indexBbIdReferenced{
		InitStatIndexToClipDB(dbId)
	}

	seeker := NewMultyDBReader(GetInitedClipStatIndexToIdentDB())
	defer seeker.Close()

	tiDB := InitCoordinatevTagToClipDB()
	iter := tiDB.DBPtr.NewIterator(nil, &tiDB.ReadOptions)

	iter.First()
	ci := 0

	statMap := imgCache.NewMyMap(true)

	var curVTagId []byte
	var curIndexAndIdent1, curIndexAndIdent2 []byte
	for iter.Valid(){
		if offset <= ci{
			if len(iter.Key()) != 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH) + CLIP_VIRTUAL_TAGID_LEN{
				continue
			}else{
				vTagAndClipIndexes := fileUtil.CopyBytesTo(iter.Key())
				fmt.Print("tag_index: ")
				fileUtil.PrintBytes(vTagAndClipIndexes)
				curVTagId = vTagAndClipIndexes[ :CLIP_VIRTUAL_TAGID_LEN]
				curIndexAndIdent1 = vTagAndClipIndexes[CLIP_VIRTUAL_TAGID_LEN : CLIP_VIRTUAL_TAGID_LEN+ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH]
				curIndexAndIdent2 = vTagAndClipIndexes[CLIP_VIRTUAL_TAGID_LEN+ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH :]

				statMap.Put(curVTagId, curIndexAndIdent1)
				statMap.Put(curVTagId, curIndexAndIdent2)

				if statMap.KeyCount() == limit + 1{
					break
				}
			}
		}
		ci ++
		iter.Next()
	}
	iter.Release()

	var curClipIndexAndIdent []byte
	vTags := statMap.KeySet()
	var clipIdent []byte
	var interfaceClipIndexes []interface{}
	var queryClipIdent []byte
	for _,vtag := range vTags{

		clipIdentMap := imgCache.NewMyMap(false)

		interfaceClipIndexes = statMap.Get(vtag)
		if 0 != len(interfaceClipIndexes){

			for _, interfaceBranchIndex := range interfaceClipIndexes {
				curClipIndexAndIdent = interfaceBranchIndex.([]byte)
				clipIdent = curClipIndexAndIdent[ImgIndex.CLIP_INDEX_BYTES_LEN:]//QueryAnyOneClipIdentByIndex(indexBbIdReferenced[0], curClipIndexAndIdent)

				queryClipIdent = testQueryAnyOneClipIdentByIndexAndIdent(indexBbIdReferenced[0], curClipIndexAndIdent)
				if !bytes.Equal(queryClipIdent, clipIdent){
					fmt.Println("find a same one")
				}

				if ImgIndex.IMG_CLIP_IDENT_LENGTH != len(clipIdent){
					fmt.Println("error, can't find index")
				}else{
					clipIdentMap.Put(clipIdent, nil)
				}
			}
		}

		fmt.Print(vtag, " --- ")
		cidents := clipIdentMap.KeySet()
		for _,cident := range cidents{
			fmt.Print(string(ImgIndex.ParseClipIdentToString(cident, "-")), " | ")
		}
		fmt.Println()
	}
}

type CoordinateClipsResult struct {
	Left, Right uint8
	Support     int
}
type CoordinateClipResultList []CoordinateClipsResult

func (this CoordinateClipResultList)Len() int {
	return len(this)
}

func (this CoordinateClipResultList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

//逆序排列
func (this CoordinateClipResultList) Less(i, j int) bool {
	return this[i].Support > this[j].Support
}

func GetCoordinateClipsFromImgIdent(imgIdent []byte, threshold int) CoordinateClipResultList {
	clipIndexMap := GetClipIndexBytesOfWhich(imgIdent[0], imgIdent, nil)
	clipIndexes := make([][]byte, config.CLIP_COUNTS_OF_IMG)
	for i:=uint8(0);i < uint8(config.CLIP_COUNTS_OF_IMG);i ++{
		clipIndexes[int(i)] = clipIndexMap[i]
	}

	return GetCoordinateClipsFromClipIndexes(clipIndexes, threshold)
}

func GetCoordinateClipsFromClipIndexes(clipIndexes [][]byte, threshold int) CoordinateClipResultList {
	var ret []CoordinateClipsResult

	for i:=uint8(0);i < config.CLIP_COUNTS_OF_IMG;i ++{
		leftClipIndex := clipIndexes[i]
		for j:=i+1;j < config.CLIP_COUNTS_OF_IMG;j ++{
			rightClipIndex := clipIndexes[j]
			support := GetCoordinateSupport(leftClipIndex, rightClipIndex)
			if support >= threshold {
				ret = append(ret, CoordinateClipsResult{Left:i, Right:j, Support:support})
			}
		}
	}
	return ret
}

func GetCoordinateSupport(leftClipIndex, rightClipIndex []byte) int {

	leftStatIndexes := ImgIndex.ClipStatIndexBranch(leftClipIndex)
	rightStatIndexes := ImgIndex.ClipStatIndexBranch(rightClipIndex)

	//键是 2 * CLIP_STAT_INDEX_BYTES_LEN
	oneHitLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH) + CLIP_VIRTUAL_TAGID_LEN + 4

	queryKey := make([]byte, ImgIndex.CLIP_STAT_INDEX_BYTES_LEN * 2)

	clipIndex1Start := 0
	clipIndex1Limit := clipIndex1Start + ImgIndex.CLIP_INDEX_BYTES_LEN

	clipIndex2Start := clipIndex1Limit + ImgIndex.IMG_CLIP_IDENT_LENGTH
	clipIndex2Limit := clipIndex2Start + ImgIndex.CLIP_INDEX_BYTES_LEN

	vtagStart := clipIndex2Limit + ImgIndex.IMG_CLIP_IDENT_LENGTH
	vtagLimit := vtagStart + CLIP_VIRTUAL_TAGID_LEN

	supportStart := oneHitLen-4
	supportLimit := supportStart + 4

	//由于 stat index 是分支的，相同的 clip index 对应了多个 stat index, 下面两个 map 用于减少匹配的次数
	leftNotSame := imgCache.NewMyMap(false)
	rightNotSame := imgCache.NewMyMap(false)

	resMap := imgCache.NewMyMap(false)

	for _,leftStatIndex := range leftStatIndexes{
		copy(queryKey, leftStatIndex)
		for _,rightStatIndex := range rightStatIndexes{
			copy(queryKey[ImgIndex.CLIP_STAT_INDEX_BYTES_LEN:], rightStatIndex)

			hits := InitCoordinateClipToVTagDB().ReadFor(queryKey)

			if 0 == len(hits){
				continue
			}

			if 0 != len(hits) % oneHitLen{
				fmt.Println("coordinate index db key length error: ", len(hits), ", not mulpty of ", oneHitLen)
				return 0
			}

			hitArray := make([][]byte, len(hits)/oneHitLen)
			ci := 0
			for i:=0;i < len(hits);i += oneHitLen{
				hitArray[ci] = fileUtil.CopyBytesTo(hits[i:i+oneHitLen])
				ci ++
			}

			for _,hit := range hitArray{
				//以 vtag 作为一键多值 map 的键, 是为了防保证每张大图计算出来的协同关系都计算一次
				vtagBytes := hit[vtagStart:vtagLimit]
				if resMap.Contains(vtagBytes){
					continue
				}

				clipIndex1 := hit[clipIndex1Start:clipIndex1Limit]
				clipIndex2 := hit[clipIndex2Start:clipIndex2Limit]

				if leftNotSame.Contains(clipIndex1){
					continue
				}
				if rightNotSame.Contains(clipIndex2){
					continue
				}

				lok := isSameClip(clipIndex1, leftClipIndex)
				rok := isSameClip(clipIndex2, rightClipIndex)
				if lok && rok{
					vtagBytes := hit[vtagStart:vtagLimit]
					supportBytes := hit[supportStart:supportLimit]
					resMap.Put(vtagBytes, supportBytes)
				}else{
					if !lok{
						leftNotSame.Put(clipIndex1, nil)
					}
					if !rok{
						rightNotSame.Put(clipIndex2, nil)
					}
				}
			}
		}

	}

	maxSupport := 0

	vtags := resMap.KeySet()
	for _,vtag := range vtags{
		interfaceSupports := resMap.Get(vtag)
		if 0 == len(interfaceSupports){
			continue
		}

		for _,isupport := range interfaceSupports{
			supportBytes := isupport.([]byte)
			curSupport := ImgIndex.BytesToInt32(supportBytes)
			if curSupport > maxSupport{
				maxSupport = curSupport
			}
		}
	}

	return maxSupport
}

func TestCoordinateIndexDBFix()  {
	FixCoordinateIndexDB()
}

func CoordinateIndexDBKeyCount()  {
	coordinateIndexDB := InitCoordinateClipToVTagDB()
	iter := coordinateIndexDB.DBPtr.NewIterator(nil, &coordinateIndexDB.ReadOptions)
	iter.First()
	count := 0

	totalValueLen := uint64(0)

	keyLen := 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
	oneHitLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH) + CLIP_VIRTUAL_TAGID_LEN + 4

	for iter.Valid(){

		if fileUtil.BytesStartWith(iter.Key(), config.STAT_KEY_PREX){
			iter.Next()
			continue
		}

		if len(iter.Key()) !=  keyLen{
			fmt.Print("error, key len is not ", keyLen, " : ")
			fileUtil.PrintBytes(iter.Key())
			iter.Next()
			continue
		}

		count ++

		if len(iter.Value()) % oneHitLen != 0{
			fmt.Print("error, value len is not mulpty of ", oneHitLen," : ", len(iter.Value()), ", key: ")
			fileUtil.PrintBytes(iter.Key())
		}
		totalValueLen += uint64(len(iter.Value()))
		iter.Next()
	}
	meanValueLen := totalValueLen/uint64(count)



	meanValueGroups := meanValueLen /uint64(oneHitLen)
	fmt.Println("coordinate index db key count: ", count, "mean of value len: ", meanValueLen,", mean of value groups: ", meanValueGroups)
	iter.Release()
}

func TestCoordinateSupport()  {
	var groups []string

	for   {
		example := "2-A0000042-4 | 2-A0000042-5"
		stdin := bufio.NewReader(os.Stdin)
		fmt.Println("input clipIdens, like: ", example, " --> ")
		var input string
		lineBytes,_,err := stdin.ReadLine()
		if nil != err{
			fmt.Println("read line error", err.Error())
			continue
		}
		input = string(lineBytes)
		if len(input) != len(example){
			fmt.Println("input length is :", len(input), ", not: ", len(example))
			continue
		}
		groups = strings.Split(input," | ")
		var leftClipIdent, rightClipIdent []byte
		var leftClipIndex, rightClipIndex []byte

		{
			leftClipIdent = parseToClipIdent(groups[0], "-")
			leftClipIndex = InitClipToIndexDB(leftClipIdent[0]).ReadFor(leftClipIdent)
		}

		{
			rightClipIdent = parseToClipIdent(groups[1], "-")
			rightClipIndex = InitClipToIndexDB(rightClipIdent[0]).ReadFor(rightClipIdent)
		}


		support := GetCoordinateSupport(leftClipIndex, rightClipIndex)

		if support > 1{
			leftImgName := string(ImgIndex.ParseImgKeyToPlainTxt(leftClipIdent[1:5]))
			rightImgName := string(ImgIndex.ParseImgKeyToPlainTxt(rightClipIdent[1:5]))
			SaveMainImgsIn([]string{leftImgName, rightImgName}, "E:/gen/coordinate/")
		}

		fmt.Println("support is : ", support)
	}
}

func TestCoordinateSupportEx()  {

	imgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	for   {
		example := "A0000042"
		stdin := bufio.NewReader(os.Stdin)
		fmt.Println("input img name, like: ", example, " --> ")
		var input string
		lineBytes,_,err := stdin.ReadLine()
		if nil != err{
			fmt.Println("read line error", err.Error())
			continue
		}
		input = string(lineBytes)
		if len(input) != len(example){
			fmt.Println("input length is :", len(input), ", not: ", len(example))
			continue
		}

		imgIdent[0] = 2
		copy(imgIdent[1:],ImgIndex.FormatImgKey([]byte(input)))

		results := GetCoordinateClipsFromImgIdent(imgIdent,2)
		if 0 == len(results){
			fmt.Println("no coordinate: ", input)
			continue
		}

		sort.Sort(results)


		fmt.Print(input, " has coordinate: ")

		for _,res := range results{
			fmt.Print("[",res.Left,"-", res.Right," : ",res.Support,"], ")
		}
		fmt.Println()

		SaveMainImgsIn([]string{input}, "E:/gen/coordinate/")
	}
}

//2-A0000042-5
func parseToClipIdent(input string, splitStr string) []byte {
	if len(input) != 12{
		return  nil
	}
	ret := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)

	groups := strings.Split(input, splitStr)
	dbId,_ := strconv.Atoi(groups[0])

	imgKey := ImgIndex.FormatImgKey([]byte(groups[1]))

	which,_ := strconv.Atoi(groups[2])

	ret[0] = uint8(dbId)
	copy(ret[1:], imgKey)
	ret[5] = uint8(which)

	return ret
}


