package dbOptions

import (
	"imgIndex"
	"fmt"
	"strconv"
	"github.com/pkg/errors"
	"imgCache"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"bytes"
	"sort"
	"config"
	"github.com/syndtr/goleveldb/leveldb/util"
	"util"
	"bufio"
	"os"
)

//
type TrainResultItem struct {
	Whiches  []uint8
	TagIndex []byte
}

type bytesList [][]byte

func (this bytesList)Len() int {
	return len(this)
}

func (this bytesList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this bytesList) Less(i, j int) bool {
	return bytes.Compare(this[i], this[j]) < 0
}

func IsValidImgId(dbId uint8) bool {
	imgDB := PickImgDB(dbId)
	if nil == imgDB{
		return false
	}
	return true
}

func GetToTrainIterator(dbId uint8) *iterator.Iterator {
	imgDB := PickImgDB(dbId)
	if nil == imgDB{
		return nil
	}
	lastDealedImgIdent := getLastTrainImgIdentOf(dbId)
	var nowBeginImgKey []byte
	history := false
	if len(lastDealedImgIdent) == 0{
		nowBeginImgKey = nil
	}else{
		nowBeginImgKey=lastDealedImgIdent[1:]	//第 0 位上是 dbId
		history = true
	}



	r := util.Range{Start: nowBeginImgKey, Limit: nil}
	iter := imgDB.DBPtr.NewIterator(&r, &imgDB.ReadOptions)
	iter.First()
	if history{
		iter.Next()	//跳过这个图
	}
	return &iter
}


func ImgRrainResultsBatchSave(dbId uint8, resultMap *imgCache.MyMap) error {

	resDB := InitImgIndexToWhichDB()
	if nil == resDB{
		return errors.New("open result db error")
	}
//	defer resDB.CloseDB()

	imgIndexDB := InitMuImgToIndexDb(dbId)
	if nil == imgIndexDB{
		return errors.New("open img index db error")
	}
//	defer imgIndexDB.CloseDB()


	clipIdentToIndexDB := InitMuClipToIndexDB(dbId)
	if nil == clipIdentToIndexDB{
		return errors.New("open clip to index db error")
	}
//	defer clipIdentToIndexDB.CloseDB()

	clipSameDB := InitClipSameDB()
	if nil == clipSameDB{
		return errors.New("open clip same db error")
	}
//	defer clipSameDB.CloseDB()

	//----------------------------

	imgIdents := resultMap.KeySet()

	if 0 == len(imgIdents){
		return nil
	}

	for _,imgIdent := range imgIdents{
		values := resultMap.Get(imgIdent)
		if 1 != len(values){
			return errors.New("which result list len is not 1:  " + getImgNamgeFromImgIdent(imgIdent))
		}

		tagIndex := values[0].(*TrainResultItem).TagIndex
		whiches := values[0].(*TrainResultItem).Whiches

		clipIndexByteOfWhich := GetClipIndexBytesOfWhich(dbId,imgIdent,whiches)
		if nil == clipIndexByteOfWhich{
			return errors.New("get clip index null: " + getImgNamgeFromImgIdent(imgIdent))
		}

		sr := imgTrainResultSave(imgIdent, clipIndexByteOfWhich, tagIndex, whiches)
		if nil != sr{
			return sr
		}
	}

	sort.Sort(bytesList(imgIdents))
	lastDealed := imgIdents[len(imgIdents) - 1]
	setLastTrainImgIdentOf(dbId, lastDealed)

	fmt.Println("lastDealed: ", getLastTrainImgIdents())

	return nil
}


/**
	格式: {分支字段|两字节统计信息|原索引数据} * 2
 */
func imgTrainResultSave(imgIdent []byte, clipIndexBytesOfWhich map[uint8] []byte, tagIndex []byte,  whiches []uint8) error {
	dbId := uint8(imgIdent[0])

	if err:=WriteImgWhiches(dbId, imgIdent, whiches); nil != err{
		return err
	}
	//该图只有一个答案, 不需要写入 clip same 库
	if 1 < len(whiches){
		WriteTheSameClips(dbId,imgIdent,clipIndexBytesOfWhich,whiches, tagIndex)
	}

	if TAG_INDEX_LENGTH == len(tagIndex){
		WriteClipTagDB(clipIndexBytesOfWhich,whiches, tagIndex)
	}
	return nil
}


var lastTrainedKeyPrefix = []byte(string(config.STAT_KEY_PREX) + "_TRAIN_LASTKEY_")
var lastTrainedKeyPrefixLimit = append(lastTrainedKeyPrefix,255)

func setLastTrainImgIdentOf(dbId uint8, lastDealed []byte)  {
	markKey := make([]byte, len(lastTrainedKeyPrefix) + 1, len(lastTrainedKeyPrefix) + 1)
	copy(markKey, lastTrainedKeyPrefix)
	markKey[len(lastTrainedKeyPrefix)] = byte(dbId)
	InitImgIndexToWhichDB().WriteTo([]byte(markKey), lastDealed)
}

func PrintResultDBStat()  {
	dealeds := getLastTrainImgIdents()
	for _, dealed:=range dealeds{
		fmt.Println(getImgNamgeFromImgIdent(dealed))
	}
}

func PrintResultDBStatOf(dbId uint8)  {
	dealed := getLastTrainImgIdentOf(dbId)
	if len(dealed) == 0{
		fmt.Println("find none")
		return
	}
	fmt.Println(getImgNamgeFromImgIdent(dealed))

}

func PrintClipSameBytes()  {
	sameDB := InitClipSameDB()
	iter := sameDB.DBPtr.NewIterator(nil, &sameDB.ReadOptions)
	iter.First()

	stdin := bufio.NewReader(os.Stdin)
	times := 0
	fmt.Print("how many times to print: ")
	fmt.Fscan(stdin, &times)

	for iter.Valid(){

		if times <=0 {
			break
		}

		curKey := iter.Key()
		fileUtil.PrintBytes(curKey[:ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_STAT_BYTES_LEN])
		fileUtil.PrintBytes(curKey[ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_STAT_BYTES_LEN : ])
		times --
		iter.Next()
	}
}

func getLastTrainImgIdents() [][]byte {
	resDB := InitImgIndexToWhichDB()
	r := util.Range{Start: []byte(lastTrainedKeyPrefix), Limit: lastTrainedKeyPrefixLimit}
	iter := resDB.DBPtr.NewIterator(&r, &resDB.ReadOptions)
	iter.First()
	var ret [][]byte
	for iter.Valid(){
		lastDealed := fileUtil.CopyBytesTo(iter.Value())
		ret = append(ret, lastDealed)
		iter.Next()
	}
	return ret
}

func getLastTrainImgIdentOf(dbId uint8) []byte {
	resDB := InitImgIndexToWhichDB()
	r := util.Range{Start: lastTrainedKeyPrefix, Limit: lastTrainedKeyPrefixLimit}
	iter := resDB.DBPtr.NewIterator(&r, &resDB.ReadOptions)
	iter.First()

	for iter.Valid(){

		curKey := iter.Key()
		if len(curKey) == 1+len(lastTrainedKeyPrefix){
			if byte(dbId) == curKey[len(curKey)-1]{
				return fileUtil.CopyBytesTo(iter.Value())
			}
		}
		iter.Next()
	}
	return nil
}



func getImgNamgeFromImgIdent (imgIdent []byte) string {
	return strconv.Itoa(int(imgIdent[0])) + "_" + string(ImgIndex.ParseImgKeyToPlainTxt(imgIdent[1:]))
}

func getClipNamgeFromImgIdent (clipIdent []byte) string {
	return strconv.Itoa(int(clipIdent[0])) + "_" + string(ImgIndex.ParseImgKeyToPlainTxt(clipIdent[1:5])) + "_" + strconv.Itoa(int(clipIdent[5]))
}