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
)

var CLIP_VIRTUAL_TAGID_LEN = 10

/*
	同时出现在多张大图中的子图. 由于同一个子图可以和不同的子图联合出现，所以计算时它将先后有不同的 virtual id
	格式: branches clipIndexBytes | virtualTagIndex --> clipIdent

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
	格式: virtualTagIndex | branches clipIndexBytes--> clipIdent
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
	coordinateCacheList.Init(true, &callBack,true, 32000)

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

func (this *ClipCoordinateCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {
	keys := kvCache.KeySet()
	if len(keys) == 0{
		return true
	}
	biToTagDB := InitClipCoordinateBranchIndexToVTagIdDB()
	tagToBiDB := InitClipCoordinatevTagIdToBranchIndexDB()

	branchIndexToVTag := leveldb.Batch{}
	vTagToBranchIndex := leveldb.Batch{}

	biToVTagBuff := make([]byte,  ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN)
	vTagToBiBuff := make([]byte,  ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN)
	//key 是 branch index | virtual tag ids , value 是 nil
	for _,key := range keys{
		if len(key) != ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN{
			continue
		}

		interfaceClipIdent := kvCache.GetValue(key)
		if len(interfaceClipIdent) == 0{
			continue
		}
		clipIdent := interfaceClipIdent[0].([]byte)

		vtagId := key[ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN :]
		branchIndex := key[: ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN]

		{
			ci := copy(biToVTagBuff, branchIndex)
			copy(biToVTagBuff[ci: ], vtagId)
		//	fmt.Print("index_tag: ")
		//	fileUtil.PrintBytes(biToVTagBuff)
			branchIndexToVTag.Put(biToVTagBuff, clipIdent)
		}

		{
			ci := copy(vTagToBiBuff, vtagId)
			copy(vTagToBiBuff[ci: ], branchIndex)
		//	fmt.Print("tag_index: ")
		//	fileUtil.PrintBytes(vTagToBiBuff)
			vTagToBranchIndex.Put(vTagToBiBuff, clipIdent)
		}
	}

	fmt.Println("write to branchIndexToVTag: ", branchIndexToVTag.Len())
	biToTagDB.WriteBatchTo(&branchIndexToVTag)
	fmt.Println("write to vTagToBranchIndex: ", vTagToBranchIndex.Len())
	tagToBiDB.WriteBatchTo(&vTagToBranchIndex)

	return true
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

	resWhiches, allBranchesIndex := SearchEx(this.visitParams.dbId, imgKey)
	var vTagId []byte
	if len(resWhiches) != 0{
		vTagId = this.visitParams.threadVirtualTagIds[uint8(visitInfo.threadId)]
	//	fmt.Print("----thread ", visitInfo.threadId, " tagId: ")
	//	fileUtil.PrintBytes(vTagId)

		for i:=0;i < len(resWhiches);i ++{
			branches := allBranchesIndex[resWhiches[i]]
			for _, branch := range branches{

				resKey := make([]byte, len(branch) + len(vTagId))
				ci := copy(resKey, branch)
				copy(resKey[ci:], vTagId)

				clipIdent := ImgIndex.GetImgClipIdent(this.visitParams.dbId,imgKey, uint8(resWhiches[i]))

				this.visitParams.cacheList.Add(visitInfo.threadId, resKey, clipIdent)
			}
		}

		fileUtil.BytesIncrement(vTagId)

	//	fmt.Print("----thread ", visitInfo.threadId, " increment: ")
	//	fileUtil.PrintBytes(this.visitParams.threadVirtualTagIds[uint8(visitInfo.threadId)])
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


//----------------------------------------------------------------------------
func VerifyCoordinateResult(indexBbIdReferenced []uint8, offset, limit int)  {

	for _,dbId := range indexBbIdReferenced{
		InitMuIndexToClipDB(dbId)
	}

	seeker := NewMultyDBReader(GetInitedClipIndexToIdentDB())

	tiDB := InitClipCoordinatevTagIdToBranchIndexDB()
	iter := tiDB.DBPtr.NewIterator(nil, &tiDB.ReadOptions)

	iter.First()
	ci := 0

	statMap := imgCache.NewMyMap(true)

	var curVTagId []byte
	var curBranchIndex []byte
	for iter.Valid(){
		if offset <= ci{
			if len(iter.Key()) != ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + CLIP_VIRTUAL_TAGID_LEN{
				continue
			}else{
				vTagAndBranch := fileUtil.CopyBytesTo(iter.Key())
				fmt.Print("tag_index: ")
				fileUtil.PrintBytes(vTagAndBranch)
				curVTagId = vTagAndBranch[ :CLIP_VIRTUAL_TAGID_LEN]
				curBranchIndex = vTagAndBranch[CLIP_VIRTUAL_TAGID_LEN :]

				statMap.Put(curVTagId, curBranchIndex)

				if statMap.KeyCount() == limit + 1{
					break
				}
			}
		}
		ci ++
		iter.Next()
	}
	iter.Release()



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