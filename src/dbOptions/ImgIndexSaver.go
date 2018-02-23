package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
	"strconv"
	"imgOptions"
	"strings"
	"config"
	"util"
	"sort"
	"imgCache"
)

type ImgIndexSaverVisitParams struct {
	dbId uint8
	cacheList imgCache.KeyValueCacheList	//缓存,key 是 imgIndex, value 是 imgIdent 列表
}
type ImgIndexSaverVisitCallBack struct {
	maxVisitCount int
	params ImgIndexSaverVisitParams
}

func (this *ImgIndexSaverVisitCallBack) GetMaxVisitCount() int{
	return this.maxVisitCount
}

func (this *ImgIndexSaverVisitCallBack) GetLastVisitPos(dbId uint8, threadId int) []byte{
	lastVisitedKey, _ := GetThreadLastDealedKey(InitMuIndexToImgDB(dbId), dbId, threadId)
	return lastVisitedKey
}

func (this *ImgIndexSaverVisitCallBack) SetMaxVisitCount(maxVisitCount int) {
	this.maxVisitCount = maxVisitCount
}

func (this *ImgIndexSaverVisitCallBack) SetParams(params *ImgIndexSaverVisitParams) {
	this.params.dbId = (*params).dbId
}

func (this *ImgIndexSaverVisitCallBack) getParams() *ImgIndexSaverVisitParams {
	return &this.params
}

func (this *ImgIndexSaverVisitCallBack) Visit(visitInfo *VisitIngInfo) bool {
	if 0 != visitInfo.curSuccessCount && 0 == visitInfo.curSuccessCount%1000{
		fmt.Println("thread ", visitInfo.threadId, " dealed ", visitInfo.curSuccessCount)
	}

	return SaveImgIndexToDBBySrcData(this.params.dbId,&this.params.cacheList,visitInfo.threadId ,visitInfo.value, visitInfo.key)
}

func (this *ImgIndexSaverVisitCallBack) VisitFinish(finishInfo *VisitFinishedInfo) {
	SetThreadLastDealedKey(InitMuIndexToImgDB(finishInfo.dbId),
		finishInfo.dbId, finishInfo.threadId,
		finishInfo.lastSuccessDealedKey,
		finishInfo.totalCount)

	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ImgIndex.ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}

//---------------------------------------------------------------------------------

type ImgIndexCacheFlushCallBack struct {
	visitor imgCache.KeyValueCacheVisitor
	dbId uint8
}

/*
func (this *ImgIndexCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {
	indexToImgBatch := leveldb.Batch{}
	imgToIndexBatch := leveldb.Batch{}



	kvCache.Visit(this.visitor,-1,[]interface{}{&indexToImgBatch, &imgToIndexBatch})

	InitMuIndexToImgDB(this.dbId).WriteBatchTo(&indexToImgBatch)
	ImgToIndexBatchSaver(this.dbId, &imgToIndexBatch)
	return true
}
*/



func (this *ImgIndexCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {

	flushSize := 6400

	indexToImgBatch := leveldb.Batch{}
	imgToIndexBatch := leveldb.Batch{}

	indexToImgDB := InitMuIndexToImgDB(this.dbId)
	imgIndexes := kvCache.KeySet()
	dbIsEmpty := indexToImgDB.IsEmpty()

	newImgIdents := make([]byte, 50, 50)	//一个 imgIdent 长度为 5

	var imgIdent []byte
	empty := []byte{}
	for _,imgIndex := range imgIndexes{

		imgIdents := kvCache.GetValue(imgIndex)

		exsitsValue := empty
		if !dbIsEmpty{
			exsitsValue = indexToImgDB.ReadFor(imgIndex)
		}


		realLen := len(exsitsValue) + len(imgIdents) * ImgIndex.IMG_IDENT_LENGTH

		//需要的内存比现有内存多，需要再分配
		if realLen > len(newImgIdents){
			newImgIdents = make([]byte, realLen)
		}

		ni :=0

		if 0 != len(exsitsValue){
			ni += copy(newImgIdents[ni:], exsitsValue)
		}

		for _,imgIdentI := range imgIdents{
			imgIdent = imgIdentI.([]byte)
			ni += copy(newImgIdents[ni:], imgIdent)

			imgToIndexBatch.Put(imgIdent, imgIndex)
		}

		if ni != realLen{
			fmt.Println("error, ni=", ni, ", realLean: ", realLen)
		}

		indexToImgBatch.Put(imgIndex, newImgIdents[: ni])

		if indexToImgBatch.Len() >= flushSize{
			InitMuIndexToImgDB(this.dbId).WriteBatchTo(&indexToImgBatch)
			indexToImgBatch.Reset()
		}
		if imgToIndexBatch.Len() >= flushSize{
			ImgToIndexBatchSaver(this.dbId, &imgToIndexBatch)
			imgToIndexBatch.Reset()
		}
	}

	if indexToImgBatch.Len() > 0{
		InitMuIndexToImgDB(this.dbId).WriteBatchTo(&indexToImgBatch)
		indexToImgBatch.Reset()
	}
	if imgToIndexBatch.Len() > 0{
		ImgToIndexBatchSaver(this.dbId, &imgToIndexBatch)
		imgToIndexBatch.Reset()
	}

	kvCache = nil

	return true
}

//------------------------------------------------------------------------------------------
type ImgIndexCacheVisitor struct {
	dbId uint8
}


//遍历 key-value, key 是 index Bytes, value 是 ident bytes(单个).
func (this *ImgIndexCacheVisitor) Visit(imgIndexBytes []byte, imgIdents []interface{},otherParams [] interface{}) bool {

	if 2 != len(otherParams){
		fmt.Println("ImgIndexCacheVisitor need 2 other params, but only: ", len(otherParams))
		return false
	}

	indexToImgBatch := otherParams[0].(*leveldb.Batch)
	imgToIndexBatch := otherParams[1].(*leveldb.Batch)

	exsitsImgIdents := InitMuIndexToImgDB(this.dbId).ReadFor(imgIndexBytes)

	if len(exsitsImgIdents) % ImgIndex.IMG_IDENT_LENGTH != 0{
		fmt.Println("exsits img ident len is not multiple of ", ImgIndex.IMG_IDENT_LENGTH)
		return false
	}

	//注意 vlist 的类型是 interface{} 数组，每一个 interface{} 实际上是 []byte
	newImgIdents := make([]byte, len(exsitsImgIdents) + ImgIndex.IMG_IDENT_LENGTH * len(imgIdents))
	ci :=0
	if 0!=len(exsitsImgIdents){
		ci += copy(newImgIdents[ci:], exsitsImgIdents)
	}
	for _,v := range imgIdents{
		var imgIdent []byte  = v.([]byte)
		ci += copy(newImgIdents[ci:], imgIdent)
		imgToIndexBatch.Put(imgIdent , imgIndexBytes)
	}
	if len(newImgIdents) % ImgIndex.IMG_IDENT_LENGTH != 0 {
		fmt.Println("new img ident len is not multiple of ", ImgIndex.IMG_IDENT_LENGTH)
		return false
	}
	indexToImgBatch.Put(imgIndexBytes, newImgIdents)

	return true
}

//------------------------------------------------------------------------------------------

func BeginImgSaveEx(dbIndex uint8, count int)  {

	//初始化 visit 的 cache, 以及 cache 满了后调用的刷新回调结构
	imgIndexCacheList := imgCache.KeyValueCacheList{}

	//缓存 img index -> img ident, 支持重复的 values
	//与 clip index saver 类似, 为了追加 value (index ident) 而不是覆盖, 需要由唯一的线程去写.
	var callBack imgCache.CacheFlushCallBack = &ImgIndexCacheFlushCallBack{visitor:&ImgIndexCacheVisitor{dbId:dbIndex}, dbId:dbIndex}
	imgIndexCacheList.Init(true, &callBack,true,640000)

	var visitCallBack VisitCallBack = &ImgIndexSaverVisitCallBack{maxVisitCount:count,
		params:ImgIndexSaverVisitParams{dbId:dbIndex, cacheList:imgIndexCacheList}}

	VisitBySeek(PickImgDB(dbIndex), visitCallBack, -1)

	//flush 剩余的 cache
	imgIndexCacheList.FlushRemainKVCaches()
	RepairTotalSize(InitMuIndexToImgDB(dbIndex))
}


func ImgIndexSaveRun(dbIndex uint8, eachThreadCount int)  {
	BeginImgSaveEx(dbIndex, eachThreadCount)
}

func SaveImgIndexToDBBySrcData(dbId uint8, cacheList *imgCache.KeyValueCacheList, threadId int, imgSrcBytes, imgKey []byte) bool {
	imgIndexBytes := GetImgIndexBySrcBytes(imgSrcBytes)
	if nil == imgIndexBytes {
		fmt.Println("get index for ", string(imgKey)," failed")
		return false
	}

	//此处添加的 value 类型是 []byte
	imgIdent := ImgIndex.GetImgIdent(dbId,imgKey)
	if len(imgIdent) % ImgIndex.IMG_IDENT_LENGTH != 0{
		fmt.Println("to save img ident len is not multiple of ", ImgIndex.IMG_IDENT_LENGTH)
	}
	cacheList.Add(threadId, imgIndexBytes, imgIdent)
	return true
}


func GetImgIndexBySrcBytes(srcData []byte) []byte {
	data := ImgOptions.FromImageFlatBytesToStructBytes(srcData)
	if nil == data{
		fmt.Println("get image struct data error")
		return nil
	}
	indexBytes := ImgIndex.GetIndexFor(data)

	return indexBytes
}

type IndexInfo struct {
	key   [] byte
	value [] byte
	coun  int
}
type IndexInfoList []IndexInfo

func (this *IndexInfo) Assign(rkey , rvalue []byte)  {
	this.key = make([]byte, len(rkey))
	this.value = make([]byte, len(rvalue))
	copy(this.key, rkey)
	copy(this.value, rvalue)

	scount := 0
	for _, v := range rvalue{
		if v==byte('-'){
			scount ++
		}
	}
	scount ++
	this.coun = scount
}

func (this IndexInfoList)Len() int {
	return len(this)
}

func (this IndexInfoList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

//逆序
func (this IndexInfoList) Less(i, j int) bool {
	return this[i].coun > this[j].coun
}

/**
	img index 库的 key - value 表示：拥有特征 key 的图像的 id ，以 - 为分隔符放在 value 中
	现在将拥有最多图像的 key 按逆序放入到 STAT_KEY_SORT_BY_VALUE_SIZE_PREX 字段中
 */
func SetIndexSortInfo(dbId uint8)  {
	imgIndexDB := InitMuIndexToImgDB(dbId)
	if nil == imgIndexDB{
		fmt.Println("open img index db failed")
		return
	}

	keyTotalSize := 0
	indexCount := GetDBTotalSize(imgIndexDB)

	if 0 == indexCount{
		fmt.Println("img index db does not contain index")
		return
	}

	indexInfoList := make([]IndexInfo, indexCount)

	iter := imgIndexDB.DBPtr.NewIterator(nil,&imgIndexDB.ReadOptions)

	i := 0
	iter.First()
	for iter.Valid(){
		if !fileUtil.BytesStartWith(iter.Key(), config.STAT_KEY_PREX){
			keyTotalSize += len(iter.Key())
			indexInfoList[i].Assign(iter.Key(), iter.Value())
			i ++
		}
		iter.Next()
	}

	sort.Sort(IndexInfoList(indexInfoList))

	//逆序将 key 分隔放入，
	//考虑到 key 是二进制的，所以分隔时要十分小心，为稳妥，分隔串不能含有重复的字节。
	// 举例说明，若分隔串为 ###，而一个 key 恰好以 # 结尾，则分隔后会导致这个 key 丢失 #，而下一个 key 会多一个 #
	//我们使用下面的字节序列去分隔
	//有 indexCount 个要分隔，则需要 indexCount -1 个, 但是循环最后会多添加一个，所以是 indexCount 个
	splitBytes := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
	keyTotalSize += (indexCount*len(splitBytes))

	res := make([]byte, keyTotalSize)

	ci := 0
	for i, indexInfo := range indexInfoList{
		if i < 2{
			fileUtil.PrintBytes(indexInfo.key)
		}
		ci += copy(res[ci:], indexInfo.key)
		ci += copy(res[ci:], splitBytes)
	}

	if ci != keyTotalSize{
		fmt.Println("calc sorted index error: ", keyTotalSize, " != ", ci)
		return
	}

	SetSortedStatInfo(InitMuIndexToImgDB(dbId), res[0: keyTotalSize-len(splitBytes)])//去掉最后一个分隔序列
}



func ReadIndexSortInfo(dbId uint8, count int){
	res := GetSortedStatInfo(InitMuIndexToImgDB(dbId))
	if nil == res{
		fmt.Println("no sorted stat info")
		return
	}

	splitBytes := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
	ids := strings.Split(string(res), string(splitBytes))

	for i, id := range  ids{
		if i == count{
			break
		}
		fileUtil.PrintBytes([]byte(id))
		DumpImagesWithImgIndex(dbId, strconv.Itoa(i) ,[]byte(id))
	}
}

func DumpImagesWithImgIndex(dbId uint8, dirName string, index []byte)  {
	indexDB := InitMuIndexToImgDB(dbId)
	if nil == indexDB{
		fmt.Println("open index db error")
		return
	}
	imgKeys ,err := indexDB.DBPtr.Get(index, &indexDB.ReadOptions)
	if err ==leveldb.ErrNotFound{
		fmt.Println("get value of index key errr")
		return
	}
	imgList := strings.Split(string(imgKeys), "-")
	SaveMainImgsIn(imgList,"E:/gen/sorted/" + dirName)
}

func DumpImageLettersWithImgIndex(dbId uint8, dirName string, index []byte)  {
	indexDB := InitMuIndexToImgDB(dbId)
	if nil == indexDB{
		fmt.Println("open index db error")
		return
	}
	imgKeys ,err := indexDB.DBPtr.Get(index, &indexDB.ReadOptions)
	if err ==leveldb.ErrNotFound{
		fmt.Println("get value of index key errr")
		return
	}
	imgList := strings.Split(string(imgKeys), "-")
	SaveMainImgsIn(imgList,"E:/gen/sorted/" + dirName)
}