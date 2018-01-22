package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/util"
	"config"
	"fmt"
	"time"
	"bufio"
	"os"
	"bytes"
	"io/ioutil"
	"strconv"
)

var readFinished chan int

func DoReadLevelDBTest()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId uint8
	var option int
	fmt.Println("input option to run(1:iterator 2:seek 0:both)")
	fmt.Fscan(stdin, &option)

	fmt.Print("input dbId to test: ")
	fmt.Fscan(stdin, &dbId)

	var callBack VisitCallBack = new (DefaultVisitCallBack)

	if option==0 || option==2{
		fmt.Println("start to do read by seek test ------")
		VisitBySeek(PickImgDB(dbId), callBack )
	}

	if option==0 || option==1{
		fmt.Println("start to do read by iterator test ------")
		ReadLevelDBByIterator(PickImgDB(dbId), config.MAX_THREAD_COUNT)
	}

	fmt.Println("all test finished----")

}

func ReadLevelDBByIterator(dbConfig *DBConfig, threadCount int)  {
	readFinished = make(chan int, threadCount)

	start := time.Now().Unix()

	for i:=0;i < threadCount;i ++{
		go ReadByIteratorOnThread(dbConfig, i)
	}

	total := 0
	for i:=0;i < threadCount; i++{
		total += <- readFinished
	}

	end := time.Now().Unix()
	fmt.Println("all finished, total: ", total, ", cost: ", (end-start), " seconds." )
}

func ReadByIteratorOnThread(dbConfig *DBConfig, threadId int)  {
	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := dbConfig.DBPtr.NewIterator(&region,&dbConfig.ReadOptions)

	count := 0
	iter.First();
	if !iter.Valid(){
		readFinished <- 0
		return
	}
	buffer := bytes.NewBufferString("")
	for iter.Valid() {
		buffer.WriteString(string(ParseImgKeyToPlainTxt(iter.Key()))+"\n")
		count ++
		iter.Next()
	}

	ioutil.WriteFile("E:/gen/iter_keys_" + strconv.Itoa(threadId), buffer.Bytes(), 0644)

	fmt.Println("thread ", threadId, " finished, count: ", count)
	readFinished <- count
}

