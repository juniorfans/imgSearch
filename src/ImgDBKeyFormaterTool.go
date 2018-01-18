package main

import (
	"dbOptions"
	"bufio"
	"os"
	"fmt"
	"strings"
	"strconv"
	"github.com/syndtr/goleveldb/leveldb/util"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
)

func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string
	var tempDBId uint8
	for  {
		fmt.Print("input db imgs to compact(split by ,): ")
		fmt.Fscan(stdin, &input)
		dbIds := strings.Split(input, ",")
		dbs := make([] *dbOptions.DBConfig, len(dbIds))

		hasError := false

		for i,_ := range dbs{

			idbId , err := (strconv.Atoi(dbIds[i]))
			if nil != err{
				fmt.Println("dbid must be int")
				hasError = true
				break
			}

			dbs[i]=dbOptions.PickImgDB(uint8(idbId))
		//	fmt.Println("start to compact db", idbId)
		//	dbs[i].DBPtr.CompactRange(util.Range{nil,nil})
		//	fmt.Println("finished compact db", idbId)
		}
		if hasError{
			continue
		}

		fmt.Print("input temp db id: ")
		fmt.Fscan(stdin, &tempDBId)
		compactDB := dbOptions.PickImgDB(tempDBId) //tmp

		if nil == compactDB{
			fmt.Println("open compact db error")
			continue
		}

		count := 0
		for _, db := range dbs{
			count += Dumpto(compactDB, db)
			db.CloseDB()
		}
		fmt.Println("total compact : ", count)
		compactDB.CloseDB()
	}

}

var keyFormatDumpFinished chan int

func Dumpto(compactDB, srcDB *dbOptions.DBConfig) int {
	maxCore := config.MAX_THREAD_COUNT
	maxCoreStr := srcDB.ReadFor(config.STAT_KEY_DOWNLOAD_MAX_CORES)
	if nil != maxCoreStr{
		maxCore, _ = strconv.Atoi(string(maxCoreStr))
	}

	keyFormatDumpFinished = make(chan int, maxCore)

	for i:=0; i<maxCore;i ++{
		go dumpCallBack(compactDB, srcDB,i, true)
	}

	count := 0
	for i:=0;i < maxCore ;i++  {
		count += (<- keyFormatDumpFinished)
	}
	return count
}

func SingleTrheadDumpto(compactDB, srcDB *dbOptions.DBConfig) int {
	maxCore := config.MAX_THREAD_COUNT
	maxCoreStr := srcDB.ReadFor(config.STAT_KEY_DOWNLOAD_MAX_CORES)
	if nil != maxCoreStr{
		maxCore, _ = strconv.Atoi(string(maxCoreStr))
	}

	count := 0
	for i:=0; i< maxCore;i ++{
		count += dumpCallBack(compactDB, srcDB,i, false)
	}

	return count
}

func dumpCallBack(compactDB, srcDB *dbOptions.DBConfig, threadId int, multyThread bool) int {
	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := srcDB.DBPtr.NewIterator(&region,&srcDB.ReadOptions)
	iter.First()
	fmt.Println("thread: ", threadId, ", begin: ", string(iter.Key()))
	count := 0
	batch := leveldb.Batch{}
	for iter.Valid(){
		if 0!=count && 0 == count % 1000{
			fmt.Println("thread ", threadId, " dealing ", count)
			compactDB.WriteBatchTo(&batch)
			batch.Reset()
		}
		newKey := dbOptions.FormatImgKey(iter.Key())
		compactDB.WriteTo(newKey, iter.Value())
		batch.Put(newKey, iter.Value())
		iter.Next()
		count ++
	}
	fmt.Println("thread: ", threadId, ", last: ", string(iter.Key()))
	fmt.Println("thread ", threadId, " finished")
	if multyThread{
		keyFormatDumpFinished <- count
	}
	return count
}