package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"util"
	"config"
	"github.com/syndtr/goleveldb/leveldb/util"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

var splitChan chan int

func MergeTo(src, target *DBConfig)  {
	splitCount := config.MAX_THREAD_COUNT

	splitChan = make(chan int, splitCount)

	iters := splitDB(src, splitCount)


	for i:=0;i < splitCount ;i ++  {
		if len(iters) >= i{
			go doMerge(iters[i], target)
		}
	}

	dealCounts := 0
	for i:=0;i < splitCount ;i ++  {
		dealCounts += <- splitChan
	}

	fmt.Println("total merge ", dealCounts)
}

func doMerge(iter iterator.Iterator, target *DBConfig)  {
	dealCounts := 0
	flushSize := 3000
	batch := leveldb.Batch{}
	for iter.Valid(){
		if config.IsValidUserDBKey(iter.Key()){
			batch.Put(iter.Key(), iter.Value())
			dealCounts ++
			if flushSize == batch.Len(){
				target.WriteBatchTo(&batch)
				batch.Reset()
			}
		}

		iter.Next()
	}
	if 0 != batch.Len(){
		target.WriteBatchTo(&batch)
		batch.Reset()
	}

	iter.Release()
	splitChan <- dealCounts
}

func TestSplitTotalCounts(db *DBConfig, splitCount int)  {
	iters := splitDB(db,splitCount)
	totalCount := 0
	for _,iter := range iters{
		totalCount += iterToEndCount(iter)
	}

	nsize := iterToEndCount( db.DBPtr.NewIterator(nil, &db.ReadOptions))
	if totalCount != nsize{
		fmt.Println("error, split total count not equals to single count, splitTotal: ", totalCount, ", singleTotal: ", nsize)
	}
}

func iterToEndCount(iter iterator.Iterator) int {
	iter.First()
	count := 0
	for iter.Valid() {
		if config.IsValidUserDBKey(iter.Key()){
			count ++
		}
		iter.Next()
	}
	return count
}

func splitDB(db *DBConfig, splitCount int) []iterator.Iterator {
	iter := db.DBPtr.NewIterator(nil, &db.ReadOptions)
	iter.First()
	if !iter.Valid(){
		return nil
	}
	nsize := iterToEndCount(iter)

	if nsize <= splitCount{
		iter.First()
		return []iterator.Iterator {iter}
	}

	ret := make([]iterator.Iterator, splitCount)
	ri:=0
	eachCount := nsize / splitCount	//最后一个划分会多一些

	iter.First()
	beginKey := fileUtil.CopyBytesTo(iter.Key())
	ci := 0
	for iter.Valid() {
		if config.IsValidUserDBKey(iter.Key()){
			if ci == eachCount{
				ci = 0
				endKey := fileUtil.CopyBytesTo(iter.Key())

				ret[ri] = db.DBPtr.NewIterator(&util.Range{Start:beginKey, Limit:endKey}, &db.ReadOptions)
				ri ++

				beginKey = fileUtil.CopyBytesTo(endKey)
			}
			ci ++	//此句放在 if 的后面是因为 util.Range 的 limit 是 exclusive 的
		}

		iter.Next();
	}

	if ri == splitCount{
		fmt.Println("split ok")
	}else if ri+1 == splitCount{
		ret[ri] = db.DBPtr.NewIterator(&util.Range{Start:beginKey, Limit:iter.Key()}, &db.ReadOptions)
	}else{
		fmt.Println("error, real split count: ", ri-1, ", > ", splitCount)
	}

	iter.Release()

	return ret
}
