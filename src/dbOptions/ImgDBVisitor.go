package dbOptions

import (
	"fmt"
	"bytes"
	"config"
	"time"
	"github.com/syndtr/goleveldb/leveldb/util"
	"util"
	"imgIndex"
)

type VisitCallBack interface {
	//回调函数 key, value 表示遍历到的键值对, curCound, curFailedCount 表示已经调用 Visit 的总次数和失败的次数.
	//返回值表示此次 Visit 的成功/失败
	Visit(*VisitIngInfo) bool

	//最多遍历多少次. 大于 0 时此配置有效否则无效
	GetMaxVisitCount() int

	//遍历完成回调函数
	VisitFinish(*VisitFinishedInfo)

	GetLastVisitPos(dbId uint8, threadId int) []byte
}

//告诉调用者当前遍历的信息
type VisitIngInfo struct {
	key, value []byte
	curCount, curSuccessCount int
	threadId int
	//visitCallBackPtr *VisitCallBack	//当前使用的 visitcallBack 指针
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

func (this *DefaultVisitCallBack) GetLastVisitPos(dbId uint8, threadId int) []byte  {
	return nil
}

func (this *DefaultVisitCallBack) Visit(*VisitIngInfo) bool {
	return true
}

func (this *DefaultVisitCallBack) VisitFinish(finishInfo *VisitFinishedInfo) {
	fmt.Println("thread ", finishInfo.threadId," dealed: ", finishInfo.totalCount ,
		", failedCount: ", (finishInfo.totalCount-finishInfo.successCount),
		", lastDealedImgKey: ", string(ImgIndex.ParseImgKeyToPlainTxt(finishInfo.lastSuccessDealedKey)))
}

var visitFinished chan int

func VisitBySeek(dbConfig *DBConfig, callBack VisitCallBack, threadCount int) int {

	//threadCount := config.MAX_THREAD_COUNT
	if threadCount < 1 || threadCount > config.MAX_THREAD_COUNT{
		threadCount = config.MAX_THREAD_COUNT
	}
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

	lastVisitPos := callback.GetLastVisitPos(dbConfig.Id, threadId)
	if 0 != len(lastVisitPos){
		iter.Seek(lastVisitPos)
		//若当前 iter 或者 Next 都是无效的, 则不再处理, 否则从 Next 处开始处理
		if !iter.Valid() || !iter.Next(){
			fmt.Println("according to last dealed key, no data to deal for thread: ", threadId)
			visitFinished <- 0
			return
		}
	}else{
		iter.Seek([]byte{threadByte})
		if !iter.Valid(){
			fmt.Println("no data for thread: ", threadId)
			visitFinished <- 0
			return
		}
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
		threadId:threadId, /*visitCallBackPtr:&callback*/}

	for {
		if !iter.Seek(imgId) || !iter.Valid(){
			copy(limitKeyExclusive, imgId)
			break
		}

		//防止顺序的 imgId 中有连续的空洞
		if bytes.Equal(imgId,iter.Key()){
//			buffer.WriteString(string(ParseImgKeyToPlainTxt(iter.Key())) + "\n")
			visitInfo.key = fileUtil.CopyBytesTo(imgId)
			visitInfo.value = fileUtil.CopyBytesTo(iter.Value())
			visitInfo.threadId = threadId
			visitInfo.curCount = totalCount
			visitInfo.curSuccessCount = successCount
			if callback.Visit(&visitInfo){
				copy(lastSuccessDealedImgId, imgId)
				successCount ++
			}
			totalCount ++
		}else{
		/*	fmt.Print("thread ",threadId," kong dong detect, imgId: ")
			fileUtil.PrintBytes(imgId)
			fmt.Print("thread ", threadId,"but iter.Key(): ", iter.Key())
			fileUtil.PrintBytes(iter.Key())
			*/
		}

		if callback.GetMaxVisitCount()>0 && totalCount >= callback.GetMaxVisitCount(){
			break
		}

		if !ImgIndex.ImgIdIncrement(imgId){
			break
		}


	}


	//ioutil.WriteFile("E:/gen/seek_keys_" + strconv.Itoa(threadId), buffer.Bytes(), 0644)

//	fmt.Println("thread ", threadId, " finished. success: ", config.ThreadIdToName[threadId], ":", successCount,
//	", totalCount: ", totalCount)

	visitFinishedInfo := VisitFinishedInfo{totalCount:totalCount, successCount:successCount,
		threadId:threadId, lastSuccessDealedKey:lastSuccessDealedImgId, dbId:dbConfig.Id}

	fmt.Println("visit for thread ", threadId, " finished, call VisitFinish")

	callback.VisitFinish(&visitFinishedInfo)
	fmt.Println("call VisitFinish end ", threadId)
	visitFinished <- successCount
	return
}
