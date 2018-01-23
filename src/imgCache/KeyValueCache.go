package imgCache

import (
	"config"
	"unsafe"
)

var TheDeleteWhenFlushCallBack DeleteWhenFlushCallBack

//一个 key 对应多个 interface{}
type KeyValueCache map[string][]interface{}
type KeyValueCacheList []*KeyValueCache

func GetKeyAsBytes(key *string) []byte {
	return *(* []byte)(unsafe.Pointer(key))
}

func GetKeyAsSringPtr(key *[]byte) *string {
	return (* string)(unsafe.Pointer(key))
}

type KVCacheFlushCallBack interface {
	//返回值表示此次 Visit 的成功/失败
	FlushKVCache(*KeyValueCacheList,int) bool

	FlushRemainKVCaches(*KeyValueCacheList) []bool

	//当为负数时不会 flush, 表示需要调用者手动 flush
	GetFlushThreshold() int
}

type DeleteWhenFlushCallBack struct {

}

//删除老的缓存条目，使用新的/触发 GC
func (this *DeleteWhenFlushCallBack)FlushKVCache (cacheList *KeyValueCacheList ,threadId int) bool {
	var curCache KeyValueCache = make(map[string] []interface{})
	(*cacheList)[threadId] =  &curCache
	//to do flush
	return true
}

//删除所有缓存条目，触发 GC
func (this *DeleteWhenFlushCallBack)FlushRemainKVCaches (cacheList *KeyValueCacheList) []bool {
	cacheList.DestroyKVCacheList()

	ret := make([]bool, config.MAX_THREAD_COUNT)
	for i:=0;i < config.MAX_THREAD_COUNT;i++{
		ret[i] = true
	}
	//to do flush
	return ret
}

func (this *KeyValueCacheList)InitKVCacheList() {
	cacheList := make([]*KeyValueCache, config.MAX_THREAD_COUNT)
	for i:=0;i < config.MAX_THREAD_COUNT;i++{
		curCache := KeyValueCache(make(map[string] []interface{}))
		cacheList[i] = &curCache
	}

	*this = cacheList
}

func (this *KeyValueCacheList)DestroyKVCacheList()  {
	for i:=0;i < config.MAX_THREAD_COUNT;i++{
		(*this)[i] = nil
	}
	*this = nil
}

func (this *KeyValueCacheList)AddToKVCache(threadId int, keyPtr *[]byte, value interface{})  {
	iss := (*string)(unsafe.Pointer(keyPtr))
	cache := (*this)[threadId]

	(*cache)[*iss] = append((*cache)[*iss], value)
}

func (this *KeyValueCacheList)AddIndexesToKVCache(threadId int, keysPtr *[][]byte, value interface{})  {
	list := *keysPtr
	keyLen := len(list)
	for i:=0;i < keyLen;i ++{
		this.AddToKVCache(threadId, &list[i], value)
	}
}

func (this *KeyValueCacheList)GetSubKVCachePtr(threadId int) *KeyValueCache {
	cache := (*this)[threadId]
	return cache
}

//缓存中有多少个条目
func (this *KeyValueCacheList)GetSubKVCacheLength(threadId int) int {
	cache := (*this)[threadId]
	return len(*cache)
}

func (this *KeyValueCacheList) FlushKVCacheIfNeed(threadId int, callBack KVCacheFlushCallBack) bool {
	if  callBack.GetFlushThreshold()>0 && this.GetSubKVCacheLength(threadId) >= callBack.GetFlushThreshold(){
		//fmt.Println("reverse clip index reach 2000, to write")
		return this.FlushKVCache(threadId, callBack)
	}
	return true
}

func (this *KeyValueCacheList) FlushKVCache(threadId int, callBack KVCacheFlushCallBack) bool {
	return (callBack).FlushKVCache(this, threadId)
}

func (this *KeyValueCacheList) FlushRemainKVCaches(callBack KVCacheFlushCallBack) []bool {
	return (callBack).FlushRemainKVCaches(this)
}