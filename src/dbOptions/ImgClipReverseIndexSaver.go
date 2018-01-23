package dbOptions

import (
	"fmt"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
	"imgOptions"
	"imgCache"
	"sync"
)

type ClipSaverVisitParams struct {
	dbId uint8
	offsetOfClip []int
	indexLength int
}

type ClipSaverVisitCallBack struct {
	params ClipSaverVisitParams
	maxVisitCount int
}

func (this *ClipSaverVisitCallBack) GetMaxVisitCount() int{
	return this.maxVisitCount
}

func (this *ClipSaverVisitCallBack) GetLastVisitPos(dbId uint8, threadId int) []byte{
	lastVisitedKey, _ := GetThreadLastDealedKey(InitImgClipsReverseIndexDB(), dbId, threadId)
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
	ret := false
	if !SaveAllClipsToDBOf(
		visitInfo.threadId,
		visitInfo.value,
		this.params.dbId,
		visitInfo.key,
		this.params.offsetOfClip,
		this.params.indexLength){
			ret = false
	}
	ret = true
	return ret
}

func (this *ClipSaverVisitCallBack) VisitFinish(finishInfo *VisitFinishedInfo) {

	SetThreadLastDealedKey(InitImgClipsReverseIndexDB(),
		finishInfo.dbId, finishInfo.threadId,
		finishInfo.lastSuccessDealedKey,
		finishInfo.totalCount)

	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}

func BeginImgClipSaveEx(dbIndex uint8, count int, offsetOfClip []int, indexLength int) {
	var visitCallBack VisitCallBack = &ClipSaverVisitCallBack{maxVisitCount:count, params:ClipSaverVisitParams{dbId:dbIndex, offsetOfClip:offsetOfClip,indexLength:indexLength}}

	flushCallBack = &ClipCacheFlushCallBack{reverseIndex:InitImgClipsReverseIndexDB()}

	//初始化线程缓存及总体缓存. 线程缓存满了之后即加入总体缓存，则总体缓存统一写入内存
	clipReverseInexCacheList.InitKVCacheList()
	collectCache = imgCache.KeyValueCache{}

	VisitBySeek(PickImgDB(dbIndex), visitCallBack)

	clipReverseInexCacheList.FlushRemainKVCaches(flushCallBack)

	RepairTotalSize(InitImgClipsReverseIndexDB())
}

func SaveAllClipsToDBOf(threadId int, srcData []byte, dbId uint8, mainImgkey []byte, offsetOfClip []int, indexLength int) bool{

	//获得 mainImgKey 的各个切图的索引数据
	indexes := GetDBIndexOfClipsBySrcData(srcData,dbId,mainImgkey,offsetOfClip, indexLength)
	if nil == indexes{
		fmt.Println("save clips to db for ", string(mainImgkey), " failed")
		return false
	}

	//保存各个索引数据
	for _, index := range indexes{
		SaveClipsToDB(threadId, &index)
	}

	//ReadValues(imgClipDB.DBPtr, 100)

	return true
}

func SaveClipsToDB(threadId int, indexData *ImgIndex.SubImgIndex) {
	//index := indexData.GetFlatInfo()
	//计算分支 index，都对应于同样的值
	indexes := indexData.GetBranchIndexBytesIn3Chanel(4, 10)
	clipReverseInexCacheList.AddIndexesToKVCache(threadId, &indexes, indexData)
	clipReverseInexCacheList.FlushKVCacheIfNeed(threadId, flushCallBack)
}

/**
	将原 oldValue 与新的 clip value 合并, 支持 oldValue 为 nil
 */
func getValueForClipsKeyEx(oldValue []byte, indexData *ImgIndex.SubImgIndex) []byte {
	if len(oldValue) % IMG_CLIP_IDENT_LENGTH != 0{
		fmt.Println("old clip ident length is not multy of ", IMG_CLIP_IDENT_LENGTH)
		return nil
	}

	if len(indexData.KeyOfMainImg) == 0{
		fmt.Println("fuck, real empty")
	}

	clipIdent := GetImgClipIdent(indexData.DBIdOfMainImg,indexData.KeyOfMainImg,indexData.Which)

	ret := make([]byte,len(oldValue)+IMG_CLIP_IDENT_LENGTH)
	ci := 0
	if 0 != len(oldValue){
		ci += copy(ret[ci:], oldValue)
	}
	ci += copy(ret[ci:], clipIdent)
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
		fmt.Println("not found image key: ", ParseImgKeyToPlainTxt(mainImgkey), err)
		return nil
	}
	return GetDBIndexOfClipsBySrcData(srcData, dbConfig.Id, mainImgkey, offsetOfClip, indexLength)
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
		fmt.Println("read jpeg data error: ", ParseImgKeyToPlainTxt(mainImgkey))
		return nil
	}

	return ImgIndex.GetClipsIndexOfImgEx(data, dbId, mainImgkey, offsetOfClip, indexLength)
}


var flushMutex sync.Mutex

var collectCache imgCache.KeyValueCache
var clipReverseInexCacheList imgCache.KeyValueCacheList
var flushCallBack imgCache.KVCacheFlushCallBack


type ClipCacheFlushCallBack struct {
	reverseIndex *DBConfig	//index 值到 clip 的映射
	index *DBConfig	//clip 到 index 的映射
}

func (this *ClipCacheFlushCallBack) GetFlushThreshold() int{
	return 2000;
}

func (this *ClipCacheFlushCallBack)FlushKVCache (cacheList *imgCache.KeyValueCacheList ,threadId int) bool {
	//to do flush
	cache := cacheList.GetSubKVCachePtr(threadId)
	if 0 == len(*cache){
		fmt.Println("thread cache is empty, no need flush: ", threadId)
		return false
	}

	AddToCollectCache(cacheList.GetSubKVCachePtr(threadId))

	return imgCache.TheDeleteWhenFlushCallBack.FlushKVCache(cacheList, threadId)
}

func (this *ClipCacheFlushCallBack)FlushRemainKVCaches (cacheList *imgCache.KeyValueCacheList) []bool {

	for i:=0;i < config.MAX_THREAD_COUNT;i ++{
		cache := cacheList.GetSubKVCachePtr(i)
		if 0 == len(*cache){
			continue
		}
		AddToCollectCache(cacheList.GetSubKVCachePtr(i))
	}

	FinishCollectCache()

	return imgCache.TheDeleteWhenFlushCallBack.FlushRemainKVCaches(cacheList)
}

//合并 cache
//map[string][]interface{}
func AddToCollectCache(kvCache *imgCache.KeyValueCache)  {
	flushMutex.Lock()

	//相同的 key, value 进行合并
	for k,vlist := range *kvCache {
		for _, v := range vlist{
			collectCache[k]/*类型是 []interface{}*/ = append(collectCache[k], v)
		}
	}

	//条目个数超过一定的数目则写数据库
	if len(collectCache) > 2000 * 16{
		fmt.Println("cache reach threshold, write to db: ", len(collectCache))
		FinishCollectCache();
	}

	flushMutex.Unlock()
}

func FinishCollectCache()  {
	reverseIndexBatch := leveldb.Batch{}
	indexBatch := leveldb.Batch{}

	//计算旧的 value
	//合并新的 value
	//写回 db
	for k,vlist := range collectCache{
		keyBytes := imgCache.GetKeyAsBytes(&k)	//转化为 bytes
		oldValue := InitImgClipsReverseIndexDB().ReadFor(keyBytes)

		for _,v := range vlist{
			var indexData *ImgIndex.SubImgIndex  = v.(*ImgIndex.SubImgIndex )
			oldValue = getValueForClipsKeyEx(oldValue, indexData)
			imgIdent := GetImgClipIdent(indexData.DBIdOfMainImg, indexData.KeyOfMainImg, indexData.Which)
			indexBatch.Put(imgIdent ,keyBytes)
		}
		if len(oldValue) % 6 !=0 {
			fmt.Println("fuck , real not multy of 6: ", len(oldValue))
		}
		reverseIndexBatch.Put(keyBytes, oldValue)
	}
	InitImgClipsReverseIndexDB().WriteBatchTo(&reverseIndexBatch)
	ImgClipsToIndexBatchSaver(&indexBatch)
	collectCache = imgCache.KeyValueCache{}	//清空缓存, 触发 GC
}