package main

import (
	"dbOptions"
	"bufio"
	"os"
	"fmt"
	"strconv"
	"github.com/syndtr/goleveldb/leveldb/util"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
)

func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId uint8
	var tempDBId uint8
	for  {
		fmt.Print("input db imgs to compact(split by ,): ")
		fmt.Fscan(stdin, &dbId)

		imgDB := dbOptions.PickImgDB(dbId)
		if !needFormat(imgDB){
			fmt.Println("current img db has been in the target format, no need to translate")
			imgDB.CloseDB()
			continue
		}

		fmt.Print("input temp db id: ")
		fmt.Fscan(stdin, &tempDBId)
		compactDB := dbOptions.PickImgDB(tempDBId) //tmp

		if nil == compactDB{
			fmt.Println("open compact db error")
			continue
		}

		count := Dumpto(compactDB, imgDB)
		fmt.Println("total compact : ", count)
		compactDB.CloseDB()
		imgDB.CloseDB()
	}

}

func needFormat(db *dbOptions.DBConfig) bool {
	firstImgKey := dbOptions.FormatImgKey([]byte("A0000001"))

	//能够查找到目标格式的 key 则已经执行过转换
	if nil != db.ReadFor(firstImgKey){
		return false
	}

	return true
}

var keyFormatDumpFinished chan int

func Dumpto(compactDB, srcDB *dbOptions.DBConfig) int {

	fmt.Println("now we repair the stat info of img db")
	total, realMaxCore := dbOptions.ImgDBStatRepair(srcDB)

	keyFormatDumpFinished = make(chan int, realMaxCore)

	for i:=0; i< int(realMaxCore);i ++{
		go dumpCallBack(compactDB, srcDB,i, true)
	}

	count := 0
	for i:=0;i < int(realMaxCore) ;i++  {
		count += (<- keyFormatDumpFinished)
	}

	if total != count {
		fmt.Println("dealed count is not equal to total, total: ", total, ", dealed: ", count)
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
	fmt.Println("thread: ", threadId, ", begin: ", dbOptions.MakeSurePlainImgIdIsOk(iter.Key()))
	count := 0
	batch := leveldb.Batch{}

	for iter.Valid(){
		if 0!=count && 0 == count % 1000{
			fmt.Println("thread ", threadId, " dealing ", count)
			compactDB.WriteBatchTo(&batch)
			batch.Reset()
		}
		newKey := dbOptions.FormatImgKey(iter.Key())
		if nil == newKey{
			continue
		}

		batch.Put(newKey, iter.Value())
		iter.Next()
		count ++
	}

	if 0 != batch.Len(){
		compactDB.WriteBatchTo(&batch)
	}

	fmt.Println("thread: ", threadId, " finished ~, last: ", dbOptions.MakeSurePlainImgIdIsOk(iter.Key()))
	if multyThread{
		keyFormatDumpFinished <- count
	}
	return count
}