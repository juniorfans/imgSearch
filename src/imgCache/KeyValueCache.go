package imgCache

import (
	"config"
	"fmt"
	"sync"
	"os"
)

/**
	一个 key-values 缓存器, 支持相同 key 存储多个 values.
	每个线程有这样的一个 key-values 缓存器
 */

//key-value 缓存. 使用组合的方式: 内部使用 MyMap
type KeyValueCache struct {
	imap *MyMap
	multyValuesSupported bool
}

func NewKeyValueCache(multyValuesSupported bool) *KeyValueCache{
	ret := KeyValueCache{}
	ret.multyValuesSupported = multyValuesSupported
	ret.imap = NewMyMap(multyValuesSupported)
	return &ret
}

func (this *KeyValueCache) IsMultyValuesEnable() bool {
	return this.multyValuesSupported
}

func (this *KeyValueCache) GetValue(key []byte) []interface{} {
	return this.imap.Get(key)
}

func (this *KeyValueCache) Add(key []byte, value interface{})  {
	this.imap.Put(key, value)
}

func (this *KeyValueCache) KeySet() [][]byte{
	return this.imap.KeySet()
}

func (this *KeyValueCache) Visit(visitor KeyValueCacheVisitor, vcount int, otherParams [] interface{}){
	var myMapVisitor MyMapVisitor = visitor
	this.imap.Visit(myMapVisitor, vcount, otherParams)
}

//合并
func (this *KeyValueCache) Merge(right *KeyValueCache)  {
	this.imap.Merge(right.imap)
}

//keyValue 缓存中 值 的个数
func (this *KeyValueCache) Size() int{
	return this.imap.GetValueCounts()
}

func (this *KeyValueCache) Destroy(){
	this.imap.Destroy()
}

//----------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------

type KeyValueCacheList struct {
	cacheList     []*KeyValueCache
	multyValuesSupported bool

	flushThroeshold int
	flushCallBack *CacheFlushCallBack
	twoLevelCache *TwoLevelKeyValueCache


}

func (this *KeyValueCacheList)Init(multyValuesSupported bool, callBack *CacheFlushCallBack, enableTwoLevelCache bool, flushThroeshold int) {
	caches := make([]*KeyValueCache, config.MAX_THREAD_COUNT)
	for i:=0;i < config.MAX_THREAD_COUNT;i++{
		caches[i] = NewKeyValueCache(multyValuesSupported)
	}
	this.cacheList = caches

	this.multyValuesSupported = multyValuesSupported
	this.flushThroeshold = flushThroeshold
	this.flushCallBack = callBack

	if enableTwoLevelCache{
		this.twoLevelCache = &TwoLevelKeyValueCache{flushCallFuncs:callBack,
			flushThroeshold:flushThroeshold*16, cache:*NewKeyValueCache(multyValuesSupported)}
	}else{
		this.twoLevelCache = nil
	}
}

func (this *KeyValueCacheList)Destroy()  {
	for i:=0;i < config.MAX_THREAD_COUNT;i++{
		this.GetSubCache(i).Destroy()
		(*this).cacheList[i] = nil
	}
	if nil != this.twoLevelCache{
		this.twoLevelCache.cache.Destroy()
	}

	this.cacheList = nil
}

func (this *KeyValueCacheList)ResetSubCache(threadId int) *KeyValueCache {
	ret := (*this).cacheList[threadId]
	(*this).cacheList[threadId] = NewKeyValueCache(ret.multyValuesSupported)
	return ret
}

func (this *KeyValueCacheList)GetSubCache(threadId int) *KeyValueCache {
	cache := (*this).cacheList[threadId]
	return cache
}

func (this *KeyValueCacheList)Add(threadId int, key []byte, value interface{})  {
	cache := this.GetSubCache(threadId)
	if nil == cache{
		fmt.Println("get sub cache null: ", threadId)
		os.Exit(-1)
	}
	cache.Add(key, value)
	this.flushIfNeed(threadId)
}

//不同的 key 相同的 values
func (this *KeyValueCacheList)AddKeysToSameValue(threadId int, keysPtr *[][]byte, value interface{})  {
	list := *keysPtr
	keyLen := len(list)
	for i:=0;i < keyLen;i ++{
		this.Add(threadId, list[i], value)
	}
	this.flushIfNeed(threadId)
}

//缓存中有多少个条目
func (this *KeyValueCacheList)Size(threadId int) int {
	return this.GetSubCache(threadId).Size()
}

func (this *KeyValueCacheList) flushIfNeed (threadId int) bool {
	if  this.flushThroeshold>0 && this.Size(threadId) >= this.flushThroeshold{
		//fmt.Println("reverse clip index reach 2000, to write")

		//若存在二级缓存则不要执行 flush 直接将缓存加入到二级缓存中
		if nil != this.twoLevelCache{
			oldCache := this.ResetSubCache(threadId)
			(*this.twoLevelCache).Add(oldCache)
			oldCache.Destroy()
			oldCache = nil
		}else{
			oldCache := this.ResetSubCache(threadId)
			(*(this.flushCallBack)).FlushCache(oldCache)
			oldCache.Destroy()
			oldCache = nil
		}
	}
	return true
}

//将所有缓存都刷新到磁盘. 非线程安全
func (this *KeyValueCacheList) FlushRemainKVCaches() {

	//若有二级缓存则将一级缓存加入到二级缓存，再作刷新
	if nil != this.twoLevelCache{
		for i,_ := range this.cacheList{
			(*this.twoLevelCache).Add(this.ResetSubCache(i))
		}
		(*this.twoLevelCache).FlushAll()
		(*this.twoLevelCache).Destroy()

	}else{
		for i,_ := range this.cacheList{
			(*(this.flushCallBack)).FlushCache(this.ResetSubCache(i))
		}
	}
}


//----------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------

type TwoLevelKeyValueCache struct {
	flushThroeshold int
	cache          KeyValueCache
	mutex          sync.Mutex
	flushCallFuncs *CacheFlushCallBack
}

func (this *TwoLevelKeyValueCache) Size() int {
	return this.cache.Size()
}

func (this *TwoLevelKeyValueCache) Reset() {
	//重置
	this.cache = *NewKeyValueCache(this.cache.multyValuesSupported)
}

func (this *TwoLevelKeyValueCache) GetFlushThreshold() int {
	return this.flushThroeshold
}
func (this *TwoLevelKeyValueCache) Add(kvCache *KeyValueCache) {
	if nil == kvCache{
		return
	}
	this.mutex.Lock()

	this.cache.Merge(kvCache)

	//条目个数超过一定的数目则写数据库
	if this.Size() > this.GetFlushThreshold(){
		fmt.Println("cache reach threshold, write to db: ", this.Size())
		this.FlushAll();
	}

	this.mutex.Unlock()
}

//刷新所有数据到磁盘. 线程安全
func (this *TwoLevelKeyValueCache) FlushAll() {
	//this.mutex.Lock() //此函数的调用者已经 lock 了，即使是同一线程，再 lock 也会陷入等待

	(*(this.flushCallFuncs)).FlushCache(&this.cache)
	this.Reset()
	//this.mutex.Unlock()
}

func (this *TwoLevelKeyValueCache) Destroy(){
	this.cache.Destroy()
}


//----------------------------------------------------------------------------------------------------------
//						user interface
//----------------------------------------------------------------------------------------------------------

type CacheFlushCallBack interface {
	//返回值表示此次 Visit 的成功/失败
	FlushCache(*KeyValueCache) bool
}



//-------------------------------------------------------------------------------------------------------
type KeyValueCacheVisitor interface {
	//若 Visit 返回 false 则停止遍历
	Visit(key []byte, values []interface{}, otherParams [] interface{}) bool
}