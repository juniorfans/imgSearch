package dbOptions

import (
	"util"
	"imgCache"
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
)

var COMBINE_CACHED_KEY_COUNT = 3000

/**
	根据 key 的前缀做值合并
	两种模式：key 的前 prefixLen 个字节作为键，后面的字节作为值。
	或者 key 的前 prefixLen 个字节作为键, db 的 value 作为值

	若键的长度不足 prefixLen 则直接写入
 */
func CombineDBByKeyPrefix(srcDB, targetDB *DBConfig, prefixLen int, valueCombineMode int, exclusivePrefi []byte)  {
	iter := srcDB.DBPtr.NewIterator(nil, &srcDB.ReadOptions)
	iter.First()

	cmap := imgCache.NewMyMap(true)
	var lastKey []byte = fileUtil.CopyBytesTo(iter.Key())
	var nkey , nvalue []byte
	for iter.Valid(){
		//检测到隔断
		if !fileUtil.BytesEqualPrefix(iter.Key(), lastKey, prefixLen){
			//可以同步
			if cmap.KeyCount() >= COMBINE_CACHED_KEY_COUNT{
				writeMapToDB(targetDB, cmap)
				cmap.Clear()
			}
		}

		if false || len(exclusivePrefi)!=0 && fileUtil.BytesStartWith(iter.Key(), exclusivePrefi){
			nkey = fileUtil.CopyBytesTo(iter.Key())
			nvalue = fileUtil.CopyBytesTo(iter.Value())
		}else{
			if 0 == valueCombineMode{	//叠加原有值
				nkey = fileUtil.CopyBytesPrefixTo(iter.Key(), prefixLen)
				nvalue = fileUtil.CopyBytesTo(iter.Value())
			}else if 1 == valueCombineMode{	//叠加 key 从 prefixLen 开始的字节作为值
				nkey = fileUtil.CopyBytesPrefixTo(iter.Key(), prefixLen)
				nvalue = fileUtil.CopyBytesSuffixTo(iter.Key(), prefixLen)
			}else {
				fmt.Println("not support value mode")
				return
			}
		}


		cmap.Put(nkey, nvalue)

		copy(lastKey, iter.Key())

		iter.Next()
	}

	if 0 != cmap.KeyCount(){
		writeMapToDB(targetDB, cmap)
		cmap.Clear()
		cmap = nil
	}
}

func writeMapToDB(targetDB *DBConfig, cmap *imgCache.MyMap) {
	batch := leveldb.Batch{}
	keys := cmap.KeySet()

	cacheLen := 4096
	valueFlatBytes := make([]byte, cacheLen)
	for _,key := range keys{
		interfaceValues := cmap.Get(key)
		if 0 == len(interfaceValues){
			continue
		}
		ci:=0
		for _, interfaceV := range interfaceValues{
			value := interfaceV.([]byte)
			if 0 == len(value){
				continue
			}
			if ci + len(value) > cacheLen {
				for ci + len(value) > cacheLen{
					cacheLen = cacheLen * 2
				}
				newCached := make([]byte, cacheLen)
				copy(newCached,valueFlatBytes[:ci])
				valueFlatBytes = newCached
			}
			ci += copy(valueFlatBytes[ci:], value)
		}
		if 0 != len(valueFlatBytes){
			batch.Put(key, valueFlatBytes[:ci])
		}
	}

	targetDB.WriteBatchTo(&batch)
}