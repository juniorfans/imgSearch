package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
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
	"github.com/syndtr/goleveldb/leveldb/util"
)

var CLIP_VIRTUAL_TAGID_LEN = 10
var MAX_CLIP_VIRTUAL_TAGID = []byte{255,255,255,255,255,255,255,255,255,255}

/*
	同时出现在多张大图中的子图. 由于同一个子图可以和不同的子图联合出现，所以计算时它将先后有不同的 virtual id
	格式: branchIndex1 | branchIndex2 | vtag --> support

	virtualTagIndex 是虚拟的 tag, 只是为了标记，无实际含义
*/
var initedClipCoordinateBranchIndexToVirtualDb map[int] *DBConfig
func InitClipCoordinateBranchIndexToVTagIdDB() *DBConfig {
	if nil == initedClipCoordinateBranchIndexToVirtualDb {
		initedClipCoordinateBranchIndexToVirtualDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipCoordinateBranchIndexToVirtualDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		if nil == retDB.initParams{
			return nil
		}
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/clip_coordinate_bindex_vtag/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_coordinate_bindex_vtag db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipCoordinateBranchIndexToVirtualDb[hash] = &retDB

	return &retDB
}

/**
	vtag | branchIndex1 | branchIndex2 --> support
 */
var initedClipCoordinateVTagIdToBranchIndexDb map[int] *DBConfig
func InitClipCoordinatevTagIdToBranchIndexDB() *DBConfig {
	if nil == initedClipCoordinateVTagIdToBranchIndexDb {
		initedClipCoordinateVTagIdToBranchIndexDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipCoordinateVTagIdToBranchIndexDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		if nil == retDB.initParams{
			return nil
		}
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/clip_coordinate_vtag_clipident/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_coordinate_vtag_clipident db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipCoordinateVTagIdToBranchIndexDb[hash] = &retDB

	return &retDB
}



//-------------------------------------------------------------------------------------------------
func ClipCoordinateLastDealedFor(threadId uint8) []byte {
	lastDealedKeyName := string(config.STAT_KEY_PREX) + "_LAST_DEALED_IMGKEY_" + strconv.Itoa(int(threadId))
	res := InitClipCoordinateBranchIndexToVTagIdDB().ReadFor([]byte(lastDealedKeyName))
	if len(res) == 0{
		return nil
	}
	return res
}

func SetClipCoordinateLastDealedFor(lastDealed []byte, threadId uint8)  {
	lastDealedKeyName := string(config.STAT_KEY_PREX) + "_LAST_DEALED_IMGKEY_" + strconv.Itoa(int(threadId))
	InitClipCoordinateBranchIndexToVTagIdDB().WriteTo([]byte(lastDealedKeyName), lastDealed)
}

func GetUnusedVirtualTagId(threadId uint8) (ret []byte , err error){
	lastUsedVirtualTagIdName := string(config.STAT_KEY_PREX) + "_UNUSED_VIRTUAL_TAG_ID_" + strconv.Itoa(int(threadId))
	res := InitClipCoordinateBranchIndexToVTagIdDB().ReadFor([]byte(lastUsedVirtualTagIdName))
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
	InitClipCoordinateBranchIndexToVTagIdDB().WriteTo([]byte(lastUsedVirtualTagIdName), lastUsed)
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
type ClipCoordinateCacheFlushCallBack struct {

}


/**
	写入协同计算结果
	branchIndex1 | branchIndex2 | vtag --> support
	vtag | branchIndex1 | branchIndex2 --> support
 */
func (this *ClipCoordinateCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {
	keys := kvCache.KeySet()
	if len(keys) == 0{
		return true
	}
	biToTagDB := InitClipCoordinateBranchIndexToVTagIdDB()
	tagToBiDB := InitClipCoordinatevTagIdToBranchIndexDB()

	branchIndexToVTag := leveldb.Batch{}
	vTagToBranchIndex := leveldb.Batch{}

	vTagToBiBuff := make([]byte,  2*ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN)
	//key 是 branch index1 | branch index2 | virtual tag ids , value 是 support
	for _,key := range keys{
		if len(key) != 2*ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN{
			fmt.Println("error, in coordinate flush cache, key length is not equal to expect: ", len(key))
			continue
		}

		interfaceSupport := kvCache.GetValue(key)
		if len(interfaceSupport) == 0{
			continue
		}

		//vtagId := key[2 * ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN :]
		//branchIndex1 := key[ : ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN]
		//branchIndex2 := key[ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN : 2*ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN]
		var supportBytes []byte
		{
			support := interfaceSupport[0].(int)
			exsitsSupport := getOldSupportFor(key[: 2 * ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN])
			if exsitsSupport > support{
				support = exsitsSupport
			}
			supportBytes = ImgIndex.Int32ToBytes(support)
		}

		branchIndexToVTag.Put(key, supportBytes)

		transformBiVtagToVtagBi(key, vTagToBiBuff)
		vTagToBranchIndex.Put(vTagToBiBuff, supportBytes)

	}

	fmt.Println("write to branchIndexToVTag: ", branchIndexToVTag.Len())
	biToTagDB.WriteBatchTo(&branchIndexToVTag)
	fmt.Println("write to vTagToBranchIndex: ", vTagToBranchIndex.Len())
	tagToBiDB.WriteBatchTo(&vTagToBranchIndex)

	return true
}

func transformBiVtagToVtagBi(bitoTagBytes []byte, buff []byte) {
	vtagId := bitoTagBytes[2 * ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN :]
	branchIndex1 := bitoTagBytes[ : ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN]
	branchIndex2 := bitoTagBytes[ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN : 2*ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN]

	ci := 0
	ci += copy(buff[ci:], vtagId)
	ci += copy(buff[ci:], branchIndex1)
	ci += copy(buff[ci:], branchIndex2)

}

/**
	在 branch index to virtual tag 库中查找已有的 branchIndex1 | branchIndex2 的记录
 */
func getOldSupportFor(branchIndexCombine []byte) (support int){
	biToTagDB := InitClipCoordinateBranchIndexToVTagIdDB()

	limit := make([]byte, len(branchIndexCombine) + CLIP_VIRTUAL_TAGID_LEN)
	copy(limit, branchIndexCombine)
	copy(limit[len(branchIndexCombine) :], MAX_CLIP_VIRTUAL_TAGID)

	fndRange := util.Range{Start:branchIndexCombine, Limit:limit}
	iter := biToTagDB.DBPtr.NewIterator(&fndRange, &biToTagDB.ReadOptions)
	iter.First()

	maxSupport := 0

	toDeleteInBiTagBatch := leveldb.Batch{}
	toDeleteInTagBiBatch := leveldb.Batch{}

	buff := fileUtil.CopyBytesTo(limit)

	fndCount := 0

	for iter.Valid(){

		supportBytes := iter.Value()
		curSupport := ImgIndex.BytesToInt32(supportBytes)
		if maxSupport < curSupport{
			maxSupport = curSupport
		}
		toDeleteInBiTagBatch.Delete(iter.Key())
		transformBiVtagToVtagBi(iter.Key(), buff)
		toDeleteInTagBiBatch.Delete(buff)

		fndCount ++

		iter.Next()
	}

	if fndCount > 1{
		fmt.Println("warning, fndCount > 1: ", fndCount)
	}

	if toDeleteInBiTagBatch.Len() != 0{
		biToTagDB := InitClipCoordinateBranchIndexToVTagIdDB()
		biToTagDB.WriteBatchTo(&toDeleteInBiTagBatch)

		tagToBiDB := InitClipCoordinatevTagIdToBranchIndexDB()
		tagToBiDB.WriteBatchTo(&toDeleteInTagBiBatch)
	}

	support = maxSupport
	return
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

	if 0 != visitInfo.curCount && visitInfo.curCount % 1000 == 0{
		fmt.Println("thread ", visitInfo.threadId, " dealing ", visitInfo.curCount)
	}

	imgKey := visitInfo.key

	whichesGroupAndCount, allBranchesIndex := SearchCoordinateForClip(this.visitParams.dbId, imgKey)
	var vTagId []byte
	if nil != whichesGroupAndCount && whichesGroupAndCount.KeyCount() != 0{
		//每个组中的 whiches 即是同时出现在某些大图中的
		vTagId = this.visitParams.threadVirtualTagIds[uint8(visitInfo.threadId)]

		theWhichesGroups := whichesGroupAndCount.KeySet()
		for _,whiches := range theWhichesGroups{
			interfaceCounts := whichesGroupAndCount.Get(whiches)
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
				leftBranches := allBranchesIndex[int(left)]
				rightBranches := allBranchesIndex[int(right)]

				for _,lb := range leftBranches{
					for _, rb := range rightBranches{
						{
							resKey := make([]byte, len(lb) + len(rb) + len(vTagId))
							ci := 0
							ci += copy(resKey[ci:], lb)
							ci += copy(resKey[ci:], rb)
							ci += copy(resKey[ci:], vTagId)

							this.visitParams.cacheList.Add(visitInfo.threadId, resKey, curGroupSupport)
						}

						{
							resKey := make([]byte, len(rb) + len(lb) + len(vTagId))
							ci := 0
							ci += copy(resKey[ci:], rb)
							ci += copy(resKey[ci:], lb)
							ci += copy(resKey[ci:], vTagId)

							this.visitParams.cacheList.Add(visitInfo.threadId, resKey, curGroupSupport)
						}
					}
				}
			}

			//for _, which := range whiches{
			//	branches := allBranchesIndex[int(which)]
			//	for _, branch := range branches{
			//
			//		resKey := make([]byte, len(branch) + len(vTagId))
			//		ci := copy(resKey, branch)
			//		copy(resKey[ci:], vTagId)
			//
			//		clipIdent := ImgIndex.GetImgClipIdent(this.visitParams.dbId,imgKey, which)
			//
			//		this.visitParams.cacheList.Add(visitInfo.threadId, resKey, clipIdent)
			//	}
			//}

			fileUtil.BytesIncrement(vTagId)
		}
	}
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

//----------------------------------------------------------------------------
func innerVerifyCoordinateResult(indexBbIdReferenced []uint8, offset, limit int)  {

	for _,dbId := range indexBbIdReferenced{
		InitClipStatIndexToIdentsDB(dbId)
	}

	seeker := NewMultyDBReader(GetInitedClipStatIndexToIdentDB())
	defer seeker.Close()

	tiDB := InitClipCoordinatevTagIdToBranchIndexDB()
	iter := tiDB.DBPtr.NewIterator(nil, &tiDB.ReadOptions)

	iter.First()
	ci := 0

	statMap := imgCache.NewMyMap(true)

	var curVTagId []byte
	var curBranchIndex1,curBranchIndex2 []byte
	for iter.Valid(){
		if offset <= ci{
			if len(iter.Key()) != 2 * ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN{
				continue
			}else{
				vTagAndBranch := fileUtil.CopyBytesTo(iter.Key())
				fmt.Print("tag_index: ")
				fileUtil.PrintBytes(vTagAndBranch)
				curVTagId = vTagAndBranch[ :CLIP_VIRTUAL_TAGID_LEN]
				curBranchIndex1 = vTagAndBranch[CLIP_VIRTUAL_TAGID_LEN : CLIP_VIRTUAL_TAGID_LEN+ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN]
				curBranchIndex2 = vTagAndBranch[CLIP_VIRTUAL_TAGID_LEN+ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN :]

				statMap.Put(curVTagId, curBranchIndex1)
				statMap.Put(curVTagId, curBranchIndex2)

				if statMap.KeyCount() == limit + 1{
					break
				}
			}
		}
		ci ++
		iter.Next()
	}
	iter.Release()


	var curBranchIndex []byte
	vTags := statMap.KeySet()
	var clipIdents [][]byte
	var interfaceClipBranchIndexes []interface{}
	for _,vtag := range vTags{

		clipIdentMap := imgCache.NewMyMap(false)

		interfaceClipBranchIndexes = statMap.Get(vtag)
		if 0 != len(interfaceClipBranchIndexes){

			for _, interfaceBranchIndex := range interfaceClipBranchIndexes{
				curBranchIndex = interfaceBranchIndex.([]byte)
				clipIdents = seeker.ReadFor(curBranchIndex)
				if 0 == len(clipIdents){
					fmt.Println("error, can't find index")
				}else{
					clipIdentMap.Put(clipIdents[0][:ImgIndex.IMG_CLIP_IDENT_LENGTH], nil)
				}

			}

		}

		fmt.Print(vtag, " --- ")
		cidents := clipIdentMap.KeySet()
		for _,cident := range cidents{
			fmt.Print(ImgIndex.ParseClipIdentToString(cident, "-"), " | ")
		}

		fmt.Println()
	}

}