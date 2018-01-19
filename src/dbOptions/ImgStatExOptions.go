package dbOptions

import (
	"fmt"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"util"
)


/**
	从 db 库中获得上一次各线程最后处理的图片的 id 和各线程处理的总图片数目 count
	id 在 db 中的键名：ZLAST_i_tid， i 是图片库 id， tid 为线程 id
	count 在 db 中的键名：ZLAST_C_i_tid， i 是图片库 id， tid 为线程 id

 */
func GetThreadLastDealedKey(db *DBConfig, dbIndex uint8, threadId int) (lastDealedKey []byte , offset int){
	key := string(config.STAT_KEY_PREX) + "_" + strconv.Itoa(int(dbIndex)) + "_" + string(config.ThreadIdToName[threadId])

	lastDealedKey, err := db.DBPtr.Get([]byte(key), &db.ReadOptions)
	if err == leveldb.ErrNotFound{
		lastDealedKey = nil
		offset = 0
		return
	}

	key = string(config.STAT_KEY_PREX) + string("_C_") + strconv.Itoa(int(dbIndex)) + "_" + string(config.ThreadIdToName[threadId])
	offsetStr, err := db.DBPtr.Get([]byte(key), &db.ReadOptions)
	if err == leveldb.ErrNotFound{
		offset = 0
		return
	}

	offset, err = strconv.Atoi(string(offsetStr))
	if err != nil{
		offset = 0
	}

	return
}


/**
	设置 db 库中各线程最后处理的图片的 id 和各线程处理的总图片数目 count
	id 在 db 中的键名：ZLAST_i_tid， i 是图片库 id， tid 为线程 id
	count 在 db 中的键名：ZLAST_C_i_tid， i 是图片库 id， tid 为线程 id

 */
func SetThreadLastDealedKey(db *DBConfig, dbIndex uint8, threadId int, lastDealedKey []byte, count int)()  {
	key := string(config.STAT_KEY_PREX) + "_" + strconv.Itoa(int(dbIndex)) + "_" +  string(config.ThreadIdToName[threadId])
	db.DBPtr.Put([]byte(key), lastDealedKey, &db.WriteOptions)

	key = string(config.STAT_KEY_PREX) + "_C_" + strconv.Itoa(int(dbIndex)) + "_" + string(config.ThreadIdToName[threadId])
	db.DBPtr.Put([]byte(key), []byte(strconv.Itoa(count)), &db.WriteOptions)
}


func RepairTotalSize(db *DBConfig) int {
	fmt.Println("repair db total size: ", db.Dir)
	if nil == db{
		fmt.Println("open img index db failed")
		return 0
	}

	iter := db.DBPtr.NewIterator(nil,&db.ReadOptions)

	indexSize := 0

	iter.First()
	for iter.Valid(){
		if !fileUtil.BytesStartWith(iter.Key(), config.STAT_KEY_PREX){
			indexSize ++
		}
		iter.Next()
	}
	iter.Release()

	fmt.Println("img index total size: ", indexSize)

	db.DBPtr.Put(config.STAT_KEY_TOTALSIZE_PREX,[]byte(strconv.Itoa(indexSize)), &db.WriteOptions)
	return indexSize
}


func GetDBTotalSize(db *DBConfig) int {
	if nil == db{
		fmt.Println("open img index db failed")
		return 0
	}

	indexSize, err := db.DBPtr.Get(config.STAT_KEY_TOTALSIZE_PREX, &db.ReadOptions)
	if err == leveldb.ErrNotFound{
		return RepairTotalSize(db)
	}else{
		ret , err := strconv.Atoi(string(indexSize))
		if err == nil{
			return ret
		}
		return RepairTotalSize(db)
	}
}

func SetSortedStatInfo(db *DBConfig, sortedStatInfo []byte)  {
	if nil == db{
		fmt.Println("open img index db failed")
		return
	}
	db.DBPtr.Put(config.STAT_KEY_SORT_BY_VALUE_SIZE_PREX,sortedStatInfo, &db.WriteOptions)
}

func GetSortedStatInfo(db *DBConfig) []byte {
	if nil == db{
		fmt.Println("open img index db failed")
		return nil
	}
	res ,err := db.DBPtr.Get(config.STAT_KEY_SORT_BY_VALUE_SIZE_PREX,&db.ReadOptions)
	if err == leveldb.ErrNotFound{
		return nil
	}
	return res
}
