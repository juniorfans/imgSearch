package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
	"imgOptions"
	"imgCache"
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
	lastVisitedKey, _ := GetThreadLastDealedKey(InitIndexToClipDB(), dbId, threadId)
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

	SetThreadLastDealedKey(InitIndexToClipDB(),
		finishInfo.dbId, finishInfo.threadId,
		finishInfo.lastSuccessDealedKey,
		finishInfo.totalCount)

	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}


type IndexToClipCacheFlushCallBack struct {

}

func (this *IndexToClipCacheFlushCallBack) FlushCache(kvCache *imgCache.KeyValueCache) bool  {
	reverseIndexBatch := leveldb.Batch{}
	indexBatch := leveldb.Batch{}

	//计算旧的 value
	//合并新的 value
	//写回 db
	for k,vlist := range kvCache.Iterator(){
		keyBytes := imgCache.GetKeyAsBytes(&k)	//转化为 bytes
		oldValue := InitIndexToClipDB().ReadFor(keyBytes)

		//注意 vlist 的类型是 interface{} 数组，每一个 interface{} 实际上是 *ImgIndex.SubImgIndex
		for _,v := range vlist{
			var indexData *ImgIndex.SubImgIndex  = v.(*ImgIndex.SubImgIndex )
			oldValue = getValueForClipsKeyEx(oldValue, indexData)
			imgIdent := GetImgClipIdent(indexData.DBIdOfMainImg, indexData.KeyOfMainImg, indexData.Which)
			indexBatch.Put(imgIdent ,keyBytes)
		}
		if len(oldValue) % 6 !=0 {
			fmt.Println("fuck , real not multiple of 6: ", len(oldValue))
		}
		reverseIndexBatch.Put(keyBytes, oldValue)
	}
	InitIndexToClipDB().WriteBatchTo(&reverseIndexBatch)
	ImgClipsToIndexBatchSaver(&indexBatch)
	return true
}


func BeginImgClipSaveEx(dbIndex uint8, count int, offsetOfClip []int, indexLength int) {
	//初始化线程缓存及总体缓存. 线程缓存满了之后即加入总体缓存，则总体缓存统一写入内存
	indexToClipCacheList := imgCache.KeyValueCacheList{}
	var callBack imgCache.CacheFlushCallBack = &IndexToClipCacheFlushCallBack{}
	indexToClipCacheList.Init(true, &callBack,true, 400)

	var visitCallBack VisitCallBack = &ClipSaverVisitCallBack{maxVisitCount:count,
		params:ClipSaverVisitParams{dbId:dbIndex, offsetOfClip:offsetOfClip,indexLength:indexLength,
			cacheList:indexToClipCacheList}}

	VisitBySeek(PickImgDB(dbIndex), visitCallBack)

	indexToClipCacheList.FlushRemainKVCaches()

	RepairTotalSize(InitIndexToClipDB())
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
		SaveClipsToDB(cacheList, threadId, &index)
	}

	return true
}

func SaveClipsToDB(cacheList *imgCache.KeyValueCacheList, threadId int, indexData *ImgIndex.SubImgIndex) {
	//index := indexData.GetFlatInfo()
	//计算分支 index，都对应于同样的值
	indexes := indexData.GetBranchIndexBytesIn3Chanel(4, 10)

	//注意, cache 的 value 的类型是 *ImgIndex.SubImgIndex
	cacheList.AddKeysToSameValue(threadId, &indexes, indexData)
}

/**
	将原 oldValue 与新的 clip value 合并, 支持 oldValue 为 nil
 */
func getValueForClipsKeyEx(oldValue []byte, indexData *ImgIndex.SubImgIndex) []byte {
	if len(oldValue) % IMG_CLIP_IDENT_LENGTH != 0{
		fmt.Println("old clip ident length is not multiple of ", IMG_CLIP_IDENT_LENGTH)
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
