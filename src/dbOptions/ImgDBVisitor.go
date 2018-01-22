package dbOptions

import (
	"fmt"
	"bytes"
	"config"
	"time"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type VisitCallBack interface {
	//回调函数 key, value 表示遍历到的键值对, curCound, curFailedCount 表示已经调用 Visit 的总次数和失败的次数.
	//返回值表示此次 Visit 的成功/失败
	Visit(*VisitIngInfo) bool

	//最多遍历多少次. 大于 0 时此配置有效否则无效
	GetMaxVisitCount() int

	//遍历完成回调函数
	VisitFinish(*VisitFinishedInfo)
}

//告诉调用者当前遍历的信息
type VisitIngInfo struct {
	key, value []byte
	curCount, curSuccessCount int
	threadId int
}

//告诉调用者当前遍历的信息
type VisitFinishedInfo struct {
	totalCount, successCount int
	threadId                 int
	lastSuccessDealedKey     []byte
	limitKey 		 []byte	//exclusive
	dbId			 uint8
}

type DefaultVisitCallBack struct {
	maxVisitCount int
}

func (this *DefaultVisitCallBack) GetMaxVisitCount() int{
	return this.maxVisitCount
}

func (this *DefaultVisitCallBack) Visit(*VisitIngInfo) bool {
	return true
}

func (this *DefaultVisitCallBack) VisitFinish(finishInfo *VisitFinishedInfo) {
	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}

var visitFinished chan int

func VisitBySeek(dbConfig *DBConfig, callBack VisitCallBack) int {

	threadCount := config.MAX_THREAD_COUNT
	visitFinished = make(chan int, threadCount)

	start := time.Now().Unix()

	for i:=0;i < threadCount;i ++{
		go visitOnThread(dbConfig, i, callBack)
	}

	total := 0
	for i:=0;i < threadCount; i++{
		total += <- visitFinished
	}

	end := time.Now().Unix()
	fmt.Println("all finished, total: ", total, ", cost: ", (end-start), " seconds." )
	return total
}

func visitOnThread(dbConfig *DBConfig, threadId int, callback VisitCallBack)  {

	if threadId > config.MAX_THREAD_COUNT-1{
		fmt.Println("threadId is too big: ", threadId)
		visitFinished <- 0
		return
	}

	db := dbConfig.DBPtr
	threadByte := config.ThreadIdToByte[threadId]

	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := db.NewIterator(&region,&dbConfig.ReadOptions)
	iter.Seek([]byte{threadByte})
	if !iter.Valid(){
		fmt.Println("no data for thread: ", threadId)
		visitFinished <- 0
		return
	}else{
		//fmt.Println("begin : ", string(ParseImgKeyToPlainTxt(iter.Key())), ", len(value): ", len(iter.Value()))
	}

//	buffer := bytes.NewBufferString("")
//	buffer.WriteString(string(ParseImgKeyToPlainTxt(iter.Key())) + "\n")

	limitKeyExclusive := make([]byte, 4)
	lastSuccessDealedImgId := make([]byte, 4)
	imgId := make([]byte, 4)
	copy(imgId, iter.Key())
	successCount := 0
	totalCount := 0
	visitInfo := VisitIngInfo{key:imgId, value:iter.Value(), curCount:totalCount, curSuccessCount: successCount,
		threadId:threadId}

	for {
		if !iter.Seek(imgId){
			copy(limitKeyExclusive, imgId)
			break
		}
		//防止顺序的 imgId 中有连续的空洞
		if bytes.Equal(imgId,iter.Key()){
//			buffer.WriteString(string(ParseImgKeyToPlainTxt(iter.Key())) + "\n")
			visitInfo.key = imgId
			visitInfo.value = iter.Value()
			visitInfo.threadId = threadId
			visitInfo.curCount = totalCount
			visitInfo.curSuccessCount = successCount
			if callback.Visit(&visitInfo){
				copy(lastSuccessDealedImgId, imgId)
				successCount ++
			}
			totalCount ++
		}

		if callback.GetMaxVisitCount()>0 && totalCount >= callback.GetMaxVisitCount(){
			break
		}

		if !ImgIdIncrement(imgId){
			break
		}
	}


	//ioutil.WriteFile("E:/gen/seek_keys_" + strconv.Itoa(threadId), buffer.Bytes(), 0644)

//	fmt.Println("thread ", threadId, " finished. success: ", config.ThreadIdToName[threadId], ":", successCount,
//	", totalCount: ", totalCount)

	visitFinishedInfo := VisitFinishedInfo{totalCount:totalCount, successCount:successCount,
		threadId:threadId, lastSuccessDealedKey:lastSuccessDealedImgId, dbId:dbConfig.Id}

	callback.VisitFinish(&visitFinishedInfo)
	visitFinished <- successCount
	return
}
