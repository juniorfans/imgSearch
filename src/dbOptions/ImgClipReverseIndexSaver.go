package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"imgOptions"
	"imgCache"
	"imgIndex"
	"config"
	"util"
)

type ClipSaverVisitParams struct {
	dbId uint8
	offsetOfClip []int
	indexLength int
	cacheList imgCache.KeyValueCacheList	//缓存,key 是 clipIndex, value 是 clipIdent 列表
}

type ClipSaverVisitCallBack struct {
	params ClipSaverVisitParams
	maxVisitCount int
}

func (this *ClipSaverVisitCallBack) GetMaxVisitCount() int{
	return this.maxVisitCount
}

func (this *ClipSaverVisitCallBack) GetLastVisitPos(dbId uint8, threadId int) []byte{
	//计算 clip index 后填入 middle 表
	lastVisitedKey, _ := GetThreadLastDealedKey(InitMuIndexToClipMiddleDB(dbId), dbId, threadId)
	return lastVisitedKey
}

func (this *ClipSaverVisitCallBack) SetMaxVisitCount(maxVisitCount int) {
	this.maxVisitCount = maxVisitCount
}

func (this *ClipSaverVisitCallBack) SetParams(params *ClipSaverVisitParams) {
	this.params.dbId = params.dbId
	this.params.offsetOfClip = params.offsetOfClip
	this.params.indexLength = params.indexLength
}

func (this *ClipSaverVisitCallBack) getParams() *ClipSaverVisitParams {
	return &this.params
}

func (this *ClipSaverVisitCallBack) Visit(visitInfo *VisitIngInfo) bool {

	if 0 != visitInfo.curSuccessCount && 0 == visitInfo.curSuccessCount%1000{
		fmt.Println("thread ", visitInfo.threadId, " dealed ", visitInfo.curSuccessCount)
	}

	return SaveAllClipsToDBOf(
		&this.params.cacheList,
		visitInfo.threadId,
		visitInfo.value,
		this.params.dbId,
		visitInfo.key,
		this.params.offsetOfClip,
		this.params.indexLength)
}

func (this *ClipSaverVisitCallBack) VisitFinish(finishInfo *VisitFinishedInfo) {

	SetThreadLastDealedKey(InitMuIndexToClipMiddleDB(finishInfo.dbId),
		finishInfo.dbId, finishInfo.threadId,
		finishInfo.lastSuccessDealedKey,
		finishInfo.totalCount)

	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ImgIndex.ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}

func BeginImgClipSaveEx(dbIndex uint8, count int, offsetOfClip []int, indexLength int) {
	//初始化线程缓存及总体缓存. 线程缓存满了之后即加入总体缓存，则总体缓存统一写入内存
	indexToClipCacheList := imgCache.KeyValueCacheList{}
	var callBack imgCache.CacheFlushCallBack = &IndexToClipCacheFlushCallBack{dbId:dbIndex}

	//二级缓存会导致性能下降: 由于使用了 mutex 去同步合并各个线程的缓存。在高密集的计算中，这个锁会导致性能下降了近一半，直观的感受是
	//cpu (8核，开启 16 个 goroutine)由 99% 占用率下降到了 50%
	//然而为了解决clip index 对应的 clip ident 追加问题，使用二级缓存，这样会使得写库时只有一个线程在写，可以边写边追加, 同时加大缓存可以减少锁的等待次数
	indexToClipCacheList.Init(true, &callBack,true, 320000)	//

	var visitCallBack VisitCallBack = &ClipSaverVisitCallBack{maxVisitCount:count,
		params:ClipSaverVisitParams{dbId:dbIndex, offsetOfClip:offsetOfClip,indexLength:indexLength,
			cacheList:indexToClipCacheList}}

	VisitBySeek(PickImgDB(dbIndex), visitCallBack, -1)

	indexToClipCacheList.FlushRemainKVCaches()

	RepairTotalSize(InitMuIndexToClipMiddleDB(dbIndex))

	//由中间表写最终结果表
	fmt.Println("start to gather result from clip_index_to_idents_middle")
	FixClipIndexToIdentDBs([]uint8{dbIndex})

	fmt.Println("start to gather result from clip_stat_index_to_idents_middle")
	FixClipStatIndexToIdentsDBs([]uint8{dbIndex})
}

func SaveAllClipsToDBOf(cacheList *imgCache.KeyValueCacheList, threadId int, srcData []byte, dbId uint8, mainImgkey []byte, offsetOfClip []int, indexLength int) bool{

	//获得 mainImgKey 的各个切图的索引数据
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,mainImgkey,offsetOfClip, indexLength)
	if nil == indexes{
		fmt.Println("save clips to db for ", string(mainImgkey), " failed")
		return false
	}

	//保存各个索引数据
	for _, index := range indexes{
	//	fmt.Println("to save: ", string(ParseImgKeyToPlainTxt(index.KeyOfMainImg)), "-", index.Which)
		SaveClipsToDB(cacheList, threadId, &index)
	}

	return true
}

func SaveClipsToDB(cacheList *imgCache.KeyValueCacheList, threadId int, theIndexData *ImgIndex.SubImgIndex) {

	sourceIndex := theIndexData.GetIndexBytesIn3Chanel()
	cacheList.Add(threadId, sourceIndex, theIndexData.ClipIdent)
	/*
	indexDataPtr := theIndexData.Clone()

	indexDataPtr.IsSourceIndex = true

	sourceIndex := indexDataPtr.GetIndexBytesIn3Chanel()
	cacheList.Add(threadId, sourceIndex, indexDataPtr)
	//注意, cache 的 value 的类型是 *ImgIndex.SubImgIndex'
	dupIndexDataPtr := indexDataPtr.Clone()
	dupIndexDataPtr.IsSourceIndex = false
	branchIndexes := dupIndexDataPtr.GetBranchIndexBytesIn3Chanel()
	for _, branchIndex := range branchIndexes{
		cacheList.Add(threadId, branchIndex, dupIndexDataPtr)
	}

	*/
}

func CalcClipCombineIndexFor(clipSourceIndex []byte) []byte {
	ret := make([]byte, ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)


	return ret
}


/**
	获得 fileName 图像中的小图
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据

	offsetOfClip 为负数则置为0，大于边界则返回 nil
	indexLength 过大或者为负数则返回从 offsetOfClip 开始的所有数据
 */
func GetDBIndexOfClips(dbConfig *DBConfig,mainImgkey []byte, offsetOfClip []int, indexLength int) []ImgIndex.SubImgIndex {
	srcData,err := dbConfig.DBPtr.Get(mainImgkey, &dbConfig.ReadOptions)
	if err == leveldb.ErrNotFound{
		fmt.Println("not found image key: ", ImgIndex.ParseImgKeyToPlainTxt(mainImgkey), err)
		return nil
	}
	return GetDBIndexOfClipsBySrcData(srcData, dbConfig.Id, mainImgkey, offsetOfClip, indexLength)
}

func QueryClipIndexesFor(dbId uint8, imgKey []byte) [] []byte {
	clipToIndexDB := InitMuClipToIndexDB(dbId)

	ret := make([] []byte, config.CLIP_COUNTS_OF_IMG)
	clipIdent := ImgIndex.GetImgClipIdent(dbId, imgKey, 0)

	for i:=0;i < int(config.CLIP_COUNTS_OF_IMG);i ++{
		curIndex := clipToIndexDB.ReadFor(clipIdent)
		if ImgIndex.CLIP_INDEX_BYTES_LEN != len(curIndex){
			fmt.Println("query clip index not exsits: ", dbId, string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)))
			return nil
		}
		ret[i] = curIndex

		fileUtil.BytesIncrement(clipIdent)
	}

	return ret
}

/**
	获得 fileName 图像中的小图
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据

	offsetOfClip 为负数则置为0，大于边界则返回 nil
	indexLength 过大或者为负数则返回从 offsetOfClip 开始的所有数据
 */
func GetDBIndexOfClipsBySrcData(srcData []byte, dbId uint8,mainImgkey []byte, offsetOfClip []int, indexLength int) []ImgIndex.SubImgIndex {

	data := ImgOptions.FromImageFlatBytesToStructBytes(srcData)

	if nil == data{
		fmt.Println("read jpeg data error: ", ImgIndex.ParseImgKeyToPlainTxt(mainImgkey))
		return nil
	}

	return ImgIndex.GetClipsIndexOfImgEx(data, dbId, mainImgkey, offsetOfClip, indexLength)
}


//--------------------------------------------------------------------------------
//遍历 cachelist 的迭代器. 暂未使用
type ClipIndexCacheVisitor struct {
	dbId uint8
}

//遍历 key-value, key 是 clip index Bytes, value 是 clip ident bytes(单个).
func (this *ClipIndexCacheVisitor) Visit(clipIndexBytes []byte, subImgIndexs []interface{},otherParams [] interface{}) bool {

	if 2 != len(otherParams){
		fmt.Println("ClipIndexCacheVisitor need 2 other params, but only: ", len(otherParams))
		return false
	}

	indexToClipBatch := otherParams[0].(*leveldb.Batch)
	clipToIndexBatch := otherParams[1].(*leveldb.Batch)

	//这里的查询非常费时间
	var exsitsClipIdents []byte = InitMuIndexToClipDB(this.dbId).ReadFor(clipIndexBytes)

	//注意 vlist 的类型是 interface{} 数组，每一个 interface{} 实际上是 *ImgIndex.SubImgInde
	for _,v := range subImgIndexs{
		var indexData *ImgIndex.SubImgIndex  = v.(*ImgIndex.SubImgIndex )

		//当前索引值若是原始索引/不是分支索引则需要写入一条 clip -> index 的记录
		if indexData.IsSourceIndex{
			clipIdent := ImgIndex.GetImgClipIdent(indexData.DBIdOfMainImg, indexData.KeyOfMainImg, indexData.Which)
			clipToIndexBatch.Put(clipIdent, clipIndexBytes)
		}
		exsitsClipIdents = mergeExsitsClipIdentAndNew(exsitsClipIdents, indexData)
	}

	if len(exsitsClipIdents) % 6 !=0 {
		fmt.Println("fuck , new clip ident length is not multiple of 6: ", len(exsitsClipIdents))
	}
	indexToClipBatch.Put(clipIndexBytes, exsitsClipIdents)

	return true
}

/**
	将原 exsitsClipInfo 与新的 clip idents 合并, 支持 exsitsClipInfo 为 nil
 */
func mergeExsitsClipIdentAndNew(exsitsClipInfo []byte, indexData *ImgIndex.SubImgIndex) []byte {
	if len(exsitsClipInfo) % ImgIndex.IMG_CLIP_IDENT_LENGTH != 0{
		fmt.Println("old clip ident length is not multiple of ", ImgIndex.IMG_CLIP_IDENT_LENGTH)
		return nil
	}

	clipIdent := ImgIndex.GetImgClipIdent(indexData.DBIdOfMainImg,indexData.KeyOfMainImg,indexData.Which)

	ret := make([]byte,len(exsitsClipInfo)+ImgIndex.IMG_CLIP_IDENT_LENGTH)
	ci := 0
	if 0 != len(exsitsClipInfo){
		ci += copy(ret[ci:], exsitsClipInfo)
	}
	ci += copy(ret[ci:], clipIdent)
	return ret
}

//-------------------------------------------------------------------------------------------
type IndexToClipCacheFlushCallBack struct {
	visitor imgCache.KeyValueCacheVisitor
	dbId uint8
}

/*
//clip index 的 kvCache 中存储的是, key: clip index bytes, values: 子图信息
func (this *IndexToClipCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {
	indexToClipBatch := leveldb.Batch{}
	clipToIndexBatch := leveldb.Batch{}

	fmt.Println("begin to visit cache")
	kvCache.Visit(this.visitor, -1, []interface{}{&indexToClipBatch, &clipToIndexBatch})
	fmt.Println("end for visit cache")

	fmt.Println("begin to flush cache to db")
	InitMuIndexToClipDB(this.dbId).WriteBatchTo(&indexToClipBatch)
	ImgClipsToIndexBatchSaver(this.dbId, &clipToIndexBatch)
	fmt.Println("begin to flush cache to db")

	indexToClipBatch.Reset()
	clipToIndexBatch.Reset()

	kvCache = nil
	return true
}

*/


/**
	写入 clip index
	键是 branchIndex|clipIdent, 值是 nil
 */
func (this *IndexToClipCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {

	flushSize := 6400
	indexToClipBatch := leveldb.Batch{}
	clipToIndexBatch := leveldb.Batch{}
	statIndexToClipIdentBatch := leveldb.Batch{}

	clipIndexes := kvCache.KeySet()	//clip indexes

	ci := 0
	fmt.Println(len(clipIndexes), " to flush")

	cacheCount := 10
	branchIndexBuffer := make([]byte, ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + cacheCount * ImgIndex.IMG_CLIP_IDENT_LENGTH)
	statBranchIndexBuffer := make([]byte, ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + cacheCount * ImgIndex.IMG_CLIP_IDENT_LENGTH)
	var clipIdentRealCount int

	for _,clipIndex := range clipIndexes{
		ci ++
		if ci % 1000 == 0{
			fmt.Println("flushing: ", ci)
		}

		interfaceClipIdents := kvCache.GetValue(clipIndex)
		if nil == interfaceClipIdents || 0 == len(interfaceClipIdents){
			continue
		}

		clipIdentRealCount = len(interfaceClipIdents)
		if  clipIdentRealCount > cacheCount{
			branchIndexBuffer = make([]byte, ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN + clipIdentRealCount * ImgIndex.IMG_CLIP_IDENT_LENGTH)
			statBranchIndexBuffer = make([]byte, ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + clipIdentRealCount * ImgIndex.IMG_CLIP_IDENT_LENGTH)

			cacheCount = clipIdentRealCount
		}

		cib := ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN
		csb := ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
		for _,v := range interfaceClipIdents {
			clipIdent := v.([]byte)
			//写入一条 clip ident -> index
			clipToIndexBatch.Put(clipIdent, clipIndex)

			//加入 clipIdent
			cib += copy(branchIndexBuffer[cib:], clipIdent)
			csb += copy(statBranchIndexBuffer[csb:], clipIdent)
		}

		//if 0 != cib % ImgIndex.IMG_CLIP_IDENT_LENGTH{
		//	fmt.Println("cib is not multipy of ",ImgIndex.IMG_CLIP_IDENT_LENGTH, " : " , cib)
		//}
		//
		//if 0 != csb % ImgIndex.IMG_CLIP_IDENT_LENGTH{
		//	fmt.Println("csb is not multipy of ",ImgIndex.IMG_CLIP_IDENT_LENGTH, " : ", csb)
		//}

		//计算分支索引
		branchIndexes := ImgIndex.ClipIndexBranch(clipIndex)
		for _,branch := range branchIndexes{
			copy(branchIndexBuffer, branch)
			indexToClipBatch.Put(branchIndexBuffer[: cib], nil)
		}


		//计算 stat index 分支索引
		statIndexes := ImgIndex.ClipStatIndexBranch(clipIndex)
		for _,branch := range statIndexes{
			copy(statBranchIndexBuffer, branch)
			statIndexToClipIdentBatch.Put(statBranchIndexBuffer[: csb], nil)
		}


		if indexToClipBatch.Len() >= flushSize{
			InitMuIndexToClipMiddleDB(this.dbId).WriteBatchTo(&indexToClipBatch)
			indexToClipBatch.Reset()
		}
		if clipToIndexBatch.Len() >= flushSize{
			ImgClipsToIndexBatchSaver(this.dbId, &clipToIndexBatch)
			clipToIndexBatch.Reset()
		}
		if statIndexToClipIdentBatch.Len() >= flushSize{
			InitClipStatIndexToIdentsMiddleDB(this.dbId).WriteBatchTo(&statIndexToClipIdentBatch)
			statIndexToClipIdentBatch.Reset()
		}
	}

	if indexToClipBatch.Len() > 0{
		InitMuIndexToClipMiddleDB(this.dbId).WriteBatchTo(&indexToClipBatch)
		indexToClipBatch.Reset()
	}
	if clipToIndexBatch.Len() > 0{
		ImgClipsToIndexBatchSaver(this.dbId, &clipToIndexBatch)
		clipToIndexBatch.Reset()
	}
	if statIndexToClipIdentBatch.Len() > 0{
		InitClipStatIndexToIdentsMiddleDB(this.dbId).WriteBatchTo(&statIndexToClipIdentBatch)
		statIndexToClipIdentBatch.Reset()
	}

	kvCache = nil
	return true
}