package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"bytes"
	"image/jpeg"
	"github.com/Comdex/imgo"
	"imgIndex"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"strconv"
	"imgOptions"
	"strings"
)



var threadMap = map[int]byte{
	0:'A',1:'B',2:'C',3:'D',4:'E',
	5:'F',6:'G',7:'H',8:'I',9:'J',
	10:'K',11:'L',12:'M',13:'N',14:'O',
	15:'P',16:'Q',17:'R',18:'S',19:'T',
	20:'U',21:'V',22:'W',23:'X',24:'Y',
	25:'Z',
}

var indexSaveFinished chan int

func GetThreadLastDealedKey(dbIndex, threadId int) (lastDealedKey []byte , offset int){
	imgIndexDB := InitImgIndexDB()

	key := string("ZLAST_") + strconv.Itoa(dbIndex) + "_" + string(threadMap[threadId])

	lastDealedKey, err := imgIndexDB.DBPtr.Get([]byte(key), &imgIndexDB.ReadOptions)
	if err == leveldb.ErrNotFound{
		lastDealedKey = nil
		offset = 0
		return
	}

	key = string("ZLAST_C_") + strconv.Itoa(dbIndex) + "_" + string(threadMap[threadId])
	offsetStr, err := imgIndexDB.DBPtr.Get([]byte(key), &imgIndexDB.ReadOptions)
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

func SetThreadLastDealedKey(dbIndex, threadId int, lastDealedKey []byte, count int)()  {
	imgIndexDB := InitImgIndexDB()

	key := string("ZLAST_") + strconv.Itoa(dbIndex) + "_" +  string(threadMap[threadId])
	imgIndexDB.DBPtr.Put([]byte(key), lastDealedKey, &imgIndexDB.WriteOptions)

	key = string("ZLAST_C_") + strconv.Itoa(dbIndex) + "_" + string(threadMap[threadId])
	imgIndexDB.DBPtr.Put([]byte(key), []byte(strconv.Itoa(count)), &imgIndexDB.WriteOptions)
}



func PrintMagicNumber(data []byte)  {
	for i:=0;i< len(data) / 8;i++{
		fmt.Printf("%X", int(data[i]))
	}
	fmt.Println()

}

func imgIndexGoUnit(dbIndex, threadId int, iter iterator.Iterator, count int)  {

	failedCount := 0
	lastDealedKey,curCount := GetThreadLastDealedKey(dbIndex, threadId)
	baseCount := 0

	threadByte := threadMap[threadId]

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
			break
		}

		if(!SaveImgIndexBySrcData(iter.Value(), iter.Key())){
			PrintMagicNumber(iter.Value())
			failedCount ++
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
	SetThreadLastDealedKey(dbIndex, threadId, lastDealedKey, curCount)
	fmt.Println("thread ", threadId, ", failedCount: ", failedCount)
	indexSaveFinished <- threadId
}

func DoImgIndexSave(dbIndex, count int) {
	cores := 8
	indexSaveFinished = make(chan int, cores)

	imgDB := InitImgDB()
	InitImgIndexDB()

	for i:=0;i != cores;i++{
		go imgIndexGoUnit(dbIndex, i, imgDB.DBPtr.NewIterator(nil,&imgDB.ReadOptions),count)
	}

	for i:=0;i < cores;i ++{
		threadId := <- indexSaveFinished
		fmt.Println("thread ", threadId ," finished")
	}
	fmt.Println("All finished ~")
}

func Stat(dbIndex int)  {
	for i:=0;i < 8;i ++ {
		_, count := GetThreadLastDealedKey(dbIndex, i)
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

func ImgIndexRepair()  {
	indexDB := InitImgIndexDB()
	if nil == indexDB{
		fmt.Println("open index db error")
		return
	}
	for threadId:=0;threadId < 16;threadId++{
		key := string("ZLAST_") + string(threadMap[threadId])
		ckey := string("ZLAST_C_") + string(threadMap[threadId])

		kvalue, err := indexDB.DBPtr.Get([]byte(key), &indexDB.ReadOptions)
		if err != leveldb.ErrNotFound{
			indexDB.DBPtr.Delete([]byte(key), &indexDB.WriteOptions)
			newKey := string("ZLAST_") + strconv.Itoa(1) + "_" +  string(threadMap[threadId])
			indexDB.DBPtr.Put([]byte(newKey), kvalue, &indexDB.WriteOptions)
		}

		ckvalue , err := indexDB.DBPtr.Get([]byte(ckey), &indexDB.ReadOptions)
		if err != leveldb.ErrNotFound{
			indexDB.DBPtr.Delete([]byte(ckey), &indexDB.WriteOptions)
			newCKey := string("ZLAST_C_") + strconv.Itoa(1) + "_" + string(threadMap[threadId])
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

func SaveImgIndexBySrcData(srcData, imgKey []byte) bool {
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

/**
	从 image 二进制字节到 image 结构化像素数据转化
 */
func FromImageFlatBytesToStructBytes(srcData []byte) [][][]uint8 {
	var reader io.Reader = bytes.NewReader([]byte(srcData))

	image, err := jpeg.Decode(reader)
	if nil != err{
		srcData = ImgOptions.FixImage(srcData)
		reader = bytes.NewReader([]byte(srcData))
		image, err = jpeg.Decode(reader)
		if nil != err{
			fmt.Println("invalid image data: ", err)
			return nil
		}
	}
	data, err := imgo.Read(image)
	if nil != err{
		fmt.Println("read jpeg data error: ", err)
		return nil
	}
	return data
}

func GetImgIndexBySrcData(srcData []byte) []byte {
	data := FromImageFlatBytesToStructBytes(srcData)
	if nil == data{
		fmt.Println("get image struct data error")
		return nil
	}
	indexBytes := ImgIndex.GetIndexFor(data)

	return indexBytes
}

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

func GetImgIndexByImgKey(imgKey []byte) []byte {
 	imgDB := InitImgDB()
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


