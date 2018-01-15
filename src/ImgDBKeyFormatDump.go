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
)

func main()  {

	stdin := bufio.NewReader(os.Stdin)
	var input string
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
		}
		if hasError{
			continue
		}

		compactDB := dbOptions.PickImgDB(254) //tmp

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
		go dumpCallBack(compactDB, srcDB,i)
	}

	count := 0
	for i:=0;i < maxCore ;i++  {
		count += (<- keyFormatDumpFinished)
	}
	return count
}

func dumpCallBack(compactDB, srcDB *dbOptions.DBConfig, threadId int) {
	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := srcDB.DBPtr.NewIterator(&region,&srcDB.ReadOptions)
	iter.First()
	fmt.Println("thread: ", threadId, ", begin: ", string(iter.Key()))
	count := 0
	for iter.Valid(){
		newKey := dbOptions.FormatImgKey(iter.Key())
		compactDB.WriteTo(newKey, iter.Value())
		iter.Next()
		count ++
	}
	fmt.Println("thread: ", threadId, ", last: ", string(iter.Key()))
	fmt.Println("thread ", threadId, " finished")
	keyFormatDumpFinished <- count
}