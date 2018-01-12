package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"strconv"
	"imgOptions"
	"strings"
	"runtime"
	"config"
	"util"
	"sort"
)





var indexSaveFinished chan int


func PrintMagicNumber(data []byte)  {
	for i:=0;i< len(data) / 8;i++{
		fmt.Printf("%X", int(data[i]))
	}
	fmt.Println()

}

func imgIndexGoUnit(dbIndex, threadId int, iter iterator.Iterator, count int)  {

	failedCount := 0
	lastDealedKey,curCount := GetThreadLastDealedKey(InitImgIndexDB(), dbIndex, threadId)
	baseCount := 0

	threadByte := config.ThreadIdToByte[threadId]

	if nil == lastDealedKey{
		iter.Seek([]byte{threadByte})
	}else{
		iter.Seek(lastDealedKey)
		iter.Next()
	}

	if !iter.Valid(){
		fmt.Println("thread ", threadId, " iterator is invalid")
		return
	}

	for iter.Valid(){
		if iter.Key()[0] != threadByte {
			//由于 level.Iterator.Key() 内部维护着存储 key 的存储空间，返回时直接返回切片
			//一旦 Iterator 向后遍历则这个存储空间中的 key 变化了，则切片的值也会变化
			iter.Prev()
			lastDealedKey = iter.Key()
			break
		}

		if(!SaveImgIndexToDBBySrcData(iter.Value(), iter.Key())){
			PrintMagicNumber(iter.Value())
			failedCount ++
			continue
		}else{

		}

		lastDealedKey = iter.Key()

		iter.Next()
		baseCount ++
		if 0 == baseCount % 1000{
			fmt.Println("thread ", threadId, " had dealed ", baseCount, ", failed count: ", failedCount)
		}
		count --
		if count <= 0{

			break
		}
	}
	curCount += baseCount
	SetThreadLastDealedKey(InitImgIndexDB(),dbIndex, threadId, lastDealedKey, curCount)
	fmt.Println("thread ", threadId, ", failedCount: ", failedCount)
	indexSaveFinished <- threadId
}

func DoImgIndexSave(dbIndex, count int) {
	cores := 8
	runtime.GOMAXPROCS(cores)

	indexSaveFinished = make(chan int, cores)

	imgDB := GetImgDBWhichPicked()
	InitImgIndexDB()

	for i:=0;i != cores;i++{
		go imgIndexGoUnit(dbIndex, i, imgDB.DBPtr.NewIterator(nil,&imgDB.ReadOptions),count)
	}

	for i:=0;i < cores;i ++{
		threadId := <- indexSaveFinished
		fmt.Println("thread ", threadId ," finished")
	}

	RepairTotalSize(InitImgIndexDB())
	fmt.Println("All finished ~")
}

func Stat(dbIndex int)  {
	for i:=0;i < 8;i ++ {
		_, count := GetThreadLastDealedKey(InitImgIndexDB(),dbIndex, i)
		fmt.Println(strconv.Itoa(count))
	}
}

func ImgIndexSaveRun(dbIndex, eachThreadCount int)  {
	imgIndexDB:=InitImgIndexDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db error")
		return
	}

	DoImgIndexSave(dbIndex, eachThreadCount)

	Stat(dbIndex)
}

func imgIndexRepair()  {
	indexDB := InitImgIndexDB()
	if nil == indexDB{
		fmt.Println("open index db error")
		return
	}
	for threadId:=0;threadId < 16;threadId++{
		key := string("ZLAST_") + string(config.ThreadIdToByte[threadId])
		ckey := string("ZLAST_C_") + string(config.ThreadIdToByte[threadId])

		kvalue, err := indexDB.DBPtr.Get([]byte(key), &indexDB.ReadOptions)
		if err != leveldb.ErrNotFound{
			indexDB.DBPtr.Delete([]byte(key), &indexDB.WriteOptions)
			newKey := string("ZLAST_") + strconv.Itoa(1) + "_" +  string(config.ThreadIdToByte[threadId])
			indexDB.DBPtr.Put([]byte(newKey), kvalue, &indexDB.WriteOptions)
		}

		ckvalue , err := indexDB.DBPtr.Get([]byte(ckey), &indexDB.ReadOptions)
		if err != leveldb.ErrNotFound{
			indexDB.DBPtr.Delete([]byte(ckey), &indexDB.WriteOptions)
			newCKey := string("ZLAST_C_") + strconv.Itoa(1) + "_" + string(config.ThreadIdToByte[threadId])
			indexDB.DBPtr.Put([]byte(newCKey), ckvalue, &indexDB.WriteOptions)
		}
	}

	iter := indexDB.DBPtr.NewIterator(nil, &indexDB.ReadOptions)
	iter.First()
	if iter.Valid(){
		iter.Seek([]byte("ZLAST_"))
		if !iter.Valid() {
			fmt.Println("find no ZLAST_")
			return
		}

		for iter.Valid(){
			k := iter.Key()
			v := iter.Value()
			if -1 ==strings.Index(string(k), "ZLAST_"){
				break;
			}
			fmt.Println(string(k), " -- ", string(v))

			iter.Next()
		}

	}else{
		fmt.Println("iterator is invalid")
	}

	iter.Release()
}

func SaveImgIndexToDBBySrcData(srcData, imgKey []byte) bool {
	imgIndexDB := InitImgIndexDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db error")
		return false
	}

	imgBytes := GetImgIndexBySrcData(srcData)
	if nil == imgBytes{
		fmt.Println("get index for ", string(imgKey)," failed")
		return false
	}

	exsitsImgKey , err := imgIndexDB.DBPtr.Get(imgBytes, &imgIndexDB.ReadOptions)
	if err != leveldb.ErrNotFound{
		newName := string(exsitsImgKey) + "-" + string(imgKey)
	//	fmt.Println(newName)
		imgKey = []byte(newName)
	}

	err = imgIndexDB.DBPtr.Put(imgBytes, imgKey, &imgIndexDB.WriteOptions)
	if nil != err{
		fmt.Println("save index error for ", string(imgKey))
		return false
	}
	return true
}


func GetImgIndexBySrcData(srcData []byte) []byte {
	data := ImgOptions.FromImageFlatBytesToStructBytes(srcData)
	if nil == data{
		fmt.Println("get image struct data error")
		return nil
	}
	indexBytes := ImgIndex.GetIndexFor(data)

	return indexBytes
}
/*
func SaveImgIndexByImgKey(imgKey []byte)  {
	imgIndexDB := InitImgIndexDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db error")
		return
	}

	imgBytes := GetImgIndexByImgKey(imgKey)
	if nil == imgBytes{
		fmt.Println("get index for ", string(imgKey)," failed")
		return
	}
	err := imgIndexDB.DBPtr.Put(imgBytes, imgKey, &imgIndexDB.WriteOptions)
	if nil != err{
		fmt.Println("save index error for ", string(imgKey))
	}
}
*/

/*
func GetImgIndexByImgKey(imgKey []byte) []byte {
 	imgDB := GetImgDBWhichPicked()
	if nil == imgDB{
		fmt.Println("open img db error")
		return nil
	}

	srcData, err := imgDB.DBPtr.Get(imgKey, &imgDB.ReadOptions)
	if err == leveldb.ErrNotFound{
		fmt.Println("can't find img: ", string(imgKey))
		return nil
	}
	return GetImgIndexBySrcData(srcData)
}
*/

type IndexInfo struct {
	key   [] byte
	value [] byte
	coun  int
}
type IndexInfoList []IndexInfo

func (this *IndexInfo) Assign(rkey , rvalue []byte)  {
	this.key = make([]byte, len(rkey))
	this.value = make([]byte, len(rvalue))
	copy(this.key, rkey)
	copy(this.value, rvalue)

	scount := 0
	for _, v := range rvalue{
		if v==byte('-'){
			scount ++
		}
	}
	scount ++
	this.coun = scount
}

func (this IndexInfoList)Len() int {
	return len(this)
}

func (this IndexInfoList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

//逆序
func (this IndexInfoList) Less(i, j int) bool {
	return this[i].coun > this[j].coun
}

/**
	img index 库的 key - value 表示：拥有特征 key 的图像的 id ，以 - 为分隔符放在 value 中
	现在将拥有最多图像的 key 按逆序放入到 STAT_KEY_SORT_BY_VALUE_SIZE_PREX 字段中
 */
func SetIndexSortInfo()  {
	imgIndexDB := InitImgIndexDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db failed")
		return
	}

	keyTotalSize := 0
	indexCount := GetDBTotalSize(imgIndexDB)

	if 0 == indexCount{
		fmt.Println("img index db does not contain index")
		return
	}

	indexInfoList := make([]IndexInfo, indexCount)

	iter := imgIndexDB.DBPtr.NewIterator(nil,&imgIndexDB.ReadOptions)

	i := 0
	iter.First()
	for iter.Valid(){
		if !fileUtil.BytesStartWith(iter.Key(), STAT_KEY_PREX){
			keyTotalSize += len(iter.Key())
			indexInfoList[i].Assign(iter.Key(), iter.Value())
			i ++
		}
		iter.Next()
	}

	sort.Sort(IndexInfoList(indexInfoList))

	//逆序将 key 分隔放入，
	//考虑到 key 是二进制的，所以分隔时要十分小心，为稳妥，分隔串不能含有重复的字节。
	// 举例说明，若分隔串为 ###，而一个 key 恰好以 # 结尾，则分隔后会导致这个 key 丢失 #，而下一个 key 会多一个 #
	//我们使用下面的字节序列去分隔
	//有 indexCount 个要分隔，则需要 indexCount -1 个, 但是循环最后会多添加一个，所以是 indexCount 个
	splitBytes := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
	keyTotalSize += (indexCount*len(splitBytes))

	res := make([]byte, keyTotalSize)

	ci := 0
	for i, indexInfo := range indexInfoList{
		if i < 2{
			PrintBytes(indexInfo.key)
		}
		ci += copy(res[ci:], indexInfo.key)
		ci += copy(res[ci:], splitBytes)
	}

	if ci != keyTotalSize{
		fmt.Println("calc sorted index error: ", keyTotalSize, " != ", ci)
		return
	}

	SetSortedStatInfo(InitImgIndexDB(), res[0: keyTotalSize-len(splitBytes)])//去掉最后一个分隔序列
}

func PrintBytes(data []byte)  {
	for _,d := range data{
		fmt.Printf("%d ", d)
	}
	fmt.Println()
}

func ReadIndexSortInfo(count int){
	res := GetSortedStatInfo(InitImgIndexDB())
	if nil == res{
		fmt.Println("no sorted stat info")
		return
	}

	splitBytes := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
	ids := strings.Split(string(res), string(splitBytes))

	for i, id := range  ids{
		if i == count{
			break
		}
		PrintBytes([]byte(id))
		DumpImagesWithImgIndex(strconv.Itoa(i) ,[]byte(id))
	}
}

func DumpImagesWithImgIndex(dirName string, index []byte)  {
	indexDB := InitImgIndexDB()
	if nil == indexDB{
		fmt.Println("open index db error")
		return
	}
	imgKeys ,err := indexDB.DBPtr.Get(index, &indexDB.ReadOptions)
	if err ==leveldb.ErrNotFound{
		fmt.Println("get value of index key errr")
		return
	}
	imgList := strings.Split(string(imgKeys), "-")
	SaveMainImgsIn(imgList,"E:/gen/sorted/" + dirName)
}

func DumpImageLettersWithImgIndex(dirName string, index []byte)  {
	indexDB := InitImgIndexDB()
	if nil == indexDB{
		fmt.Println("open index db error")
		return
	}
	imgKeys ,err := indexDB.DBPtr.Get(index, &indexDB.ReadOptions)
	if err ==leveldb.ErrNotFound{
		fmt.Println("get value of index key errr")
		return
	}
	imgList := strings.Split(string(imgKeys), "-")
	SaveMainImgsIn(imgList,"E:/gen/sorted/" + dirName)
}