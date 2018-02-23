package main

import (
	"dbOptions"
	"bufio"
	"os"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"config"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
)

/**
	将 img 库的 key 变换.
 */
func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId uint8
	var tempDBId, newThreadCount uint8
	for  {
		fmt.Print("input db imgs to compact(split by ,): ")
		fmt.Fscan(stdin, &dbId)

		imgDB := dbOptions.PickImgDB(dbId)

		fmt.Print("input temp db id and newThreadCount: ")
		fmt.Fscan(stdin, &tempDBId, &newThreadCount)
		compactDB := dbOptions.PickImgDB(tempDBId) //tmp

		if nil == compactDB{
			fmt.Println("open compact db error")
			continue
		}

		count := Splitto(newThreadCount, compactDB, imgDB)
		fmt.Println("total compact : ", count)
		compactDB.CloseDB()
	}

}

var keyFormatDumpFinished chan int

func Splitto(newThreadCount uint8, compactDB, srcDB *dbOptions.DBConfig) int {

	fmt.Println("now we repair the stat info of img db")
	total, realMaxCore := dbOptions.ImgDBStatRepair(srcDB)

	if realMaxCore == newThreadCount{
		fmt.Println("current max core has been the: ", newThreadCount, ", no need to translate")
		return total
	}

	keyFormatDumpFinished = make(chan int, realMaxCore)
	var i uint8 = 0
	for; i<realMaxCore;i ++ {
		go splitCallBack(newThreadCount, compactDB, srcDB, int(i))
	}

	count := 0
	i = 0
	for ;i < realMaxCore ;i++  {
		count += (<- keyFormatDumpFinished)
	}
	if total != count {
		fmt.Println("dealed count is not equal to total, total: ", total, ", dealed: ", count)
	}

	total, realMaxCore = dbOptions.ImgDBStatRepair(srcDB)
	fmt.Println("after transalte, total: ", total, ", max cores: ", newThreadCount)

	return count
}

func splitCallBack(newThreadCount uint8,compactDB, srcDB *dbOptions.DBConfig, threadId int) int {

	total ,maxCores,_,_,_,_ := dbOptions.GetStatInfo(srcDB)

	if maxCores > config.MAX_THREAD_COUNT/2{
		fmt.Println("current thread count is too big, not allow to double")
		return 0
	}

	fmt.Println("current maxCores: ", maxCores, ", total: ", total)
	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := srcDB.DBPtr.NewIterator(&region,&srcDB.ReadOptions)
	iter.First()
	fmt.Println("thread: ", threadId, ", begin: ", string(ImgIndex.FormatImgKey(iter.Key())))

	eachThreadCount := total / int(newThreadCount)
	fmt.Println("each thread new count: ", eachThreadCount)

	base := config.ThreadIdToByte[0]

	count := 0
	ci := 0

	batch := leveldb.Batch{}
	tranCount := 0

	newKey := make([]byte, 8)

	for iter.Valid(){
		ci ++
		copy(newKey, iter.Key())
		//原每个线程所属数据一半扩增
		if ci >= eachThreadCount{
			oldThreadId := int(newKey[0])-int(base)
			newKey[0] = config.ThreadIdToByte[oldThreadId + maxCores]

			tranCount ++
		}

		batch.Put(newKey[0: len(iter.Key())], iter.Value())
		count ++

		if 0!=count && 0 == count % 1000{
			compactDB.WriteBatchTo(&batch)
			batch.Reset()
			fmt.Println("thread ", threadId, " dealing ", count)
		}
		iter.Next()
	}

	if 0 != batch.Len(){
		compactDB.WriteBatchTo(&batch)
	}

	fmt.Println("thread ", threadId, " finished~ translate Count : ", tranCount, ", totalCount: ", count)

	keyFormatDumpFinished <- count
	return count
}