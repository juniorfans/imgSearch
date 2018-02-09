package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
	"errors"
	"config"
	"util"
	"github.com/syndtr/goleveldb/leveldb/util"
)


func GetClipIndexBytesOfWhich(dbId uint8, imgIdent []byte, whiches []uint8) map[uint8] []byte {
	clipIdentToIndexDB := InitMuClipToIndexDB(dbId)

	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIdent, imgIdent)

	clipIndexes := make(map[uint8] []byte)
	for _,which := range whiches{
		clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = byte(which)
		curIndex := clipIdentToIndexDB.ReadFor(clipIdent)
		if 0 == len(curIndex){
			fmt.Println("get clip index null: ", getClipNamgeFromImgIdent(clipIdent))
			return nil
		}
		clipIndexes[which] = curIndex
	}
	return clipIndexes
}


/*
	主题相似的子图
	格式: (branches clipIndexBytes | branches clipIndexBytes) ---> tagIndex
*/
var initedClipSameDb map[int] *DBConfig
func InitClipSameDB() *DBConfig {
	if nil == initedClipSameDb{
		initedClipSameDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipSameDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		if nil == retDB.initParams{
			return nil
		}
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/clip_to_same_clip/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_to_same_clip db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipSameDb[hash] = &retDB

	return &retDB
}

func WriteTheSameClips(dbId uint8, imgIdent []byte, clipIndexBytesOfWhich map[uint8] []byte, whiches []uint8, tagIndex []byte)  {
	clipSameDB := InitClipSameDB()

	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIdent, imgIdent)

	sameBatch := leveldb.Batch{}
	branchLen := ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_STAT_BYTES_LEN
	toAddKey := make([]byte, 2*branchLen)
	toDupKey := make([]byte, 2*branchLen)

	for i:=0;i < len(whiches);i ++{
		iw := whiches[i]
		iIndex := clipIndexBytesOfWhich[iw]
		iBranches := ImgIndex.ClipIndexBranch(iIndex)
		for _, iBranch := range iBranches{
			copy(toAddKey, iBranch)
			for j:=i+1;j < len(whiches);j ++{
				jw := whiches[j]
				jIndex := clipIndexBytesOfWhich[jw]

				jBranches := ImgIndex.ClipIndexBranch(jIndex)
				for _,jBranch := range jBranches{
					copy(toAddKey[branchLen:], jBranch)
					sameBatch.Put(toAddKey, tagIndex)

					//倒置 toAddKey
					copy(toDupKey, toAddKey[branchLen: ])
					copy(toDupKey[branchLen: ], toAddKey[:branchLen])
					sameBatch.Put(toDupKey, tagIndex)
				}
			}
		}
	}

	clipSameDB.WriteBatchTo(&sameBatch)
}

//---------------------------------------------------------------------------
/*img 被选择的子图
	格式: (img source index bytes) --> which bytes
*/
var initedImgWhichesDb map[int] *DBConfig
func InitImgIndexToWhichDB() *DBConfig {
	if nil == initedImgWhichesDb {
		initedImgWhichesDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedImgWhichesDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/img_index_to_whiches/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this img result db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedImgWhichesDb[hash] = &retDB

	return &retDB
}

func WriteImgWhiches(dbId uint8, imgIdent []byte, whiches []uint8) error {
	//写入 img index ---> whiches
	resDB := InitImgIndexToWhichDB()

	imgIndexDB := InitMuImgToIndexDb(dbId)

	index := imgIndexDB.ReadFor(imgIdent)
	if 0 == len(index){
		fmt.Println("get img index null: ", getImgNamgeFromImgIdent(imgIdent))
		return errors.New("get img index null: " + getImgNamgeFromImgIdent(imgIdent))
	}
	//[]uint8 造价于 []byte
	resDB.WriteTo(index, whiches)
	return nil
}

//----------------------------------------------------------------
/*
	给 clip 打标签
	branches clipIndex --> tag
*/

var initedClipIndexToTagDb map[int] *DBConfig

func InitClipIndexToTagDB() *DBConfig {
	if nil == initedClipIndexToTagDb {
		initedClipIndexToTagDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipIndexToTagDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/clip_index_to_tag/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_index_to_tag db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipIndexToTagDb[hash] = &retDB

	return &retDB
}

//--------------------------------------------------------------
/**
	各个 tag 与哪些 clipIndex 关联
	格式: (tag| branches clipindex) --> nil
 */
var initedTagToClipIndexDb map[int] *DBConfig

func InitTagToClipIndexDB() *DBConfig {
	if nil == initedTagToClipIndexDb {
		initedTagToClipIndexDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagToClipIndexDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/tag_to_clip_index/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_to_clip_index db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagToClipIndexDb[hash] = &retDB

	return &retDB
}

/**
	写入 bracnes clipIndex --> tagId 和 (tagId | branches clipindex) --> nil

 */
func WriteClipTagDB(clipIndexBytesOfWhich map[uint8] []byte, whiches []uint8, tagIndex []byte)  {
	tagToBranchesIndexBatch := leveldb.Batch{}
	branchesIndexToTagIdBatch := leveldb.Batch{}
	branchLen := ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_STAT_BYTES_LEN
	toAddKey := make([]byte, branchLen + TAG_INDEX_LENGTH)
	copy(toAddKey[0:TAG_INDEX_LENGTH], tagIndex)

	for i:=0;i < len(whiches);i ++{
		iw := whiches[i]
		iIndex := clipIndexBytesOfWhich[iw]
		iBranches := ImgIndex.ClipIndexBranch(iIndex)
		for _, iBranch := range iBranches{
			branchesIndexToTagIdBatch.Put(iBranch, tagIndex)
			copy(toAddKey[TAG_INDEX_LENGTH:], iBranch)

			tagToBranchesIndexBatch.Put(toAddKey, nil)
		}
	}

	InitTagToClipIndexDB().WriteBatchTo(&tagToBranchesIndexBatch)
	InitClipIndexToTagDB().WriteBatchTo(&branchesIndexToTagIdBatch)
}





//--------------------------------------------------------------
/**
	tag 库
	格式: tag id(2 字节长度) --> tag name
 */
var initedTagIndexToNameDb map[int] *DBConfig

func InitTagIndexToNameDB() *DBConfig {
	if nil == initedTagIndexToNameDb {
		initedTagIndexToNameDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagIndexToNameDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/tag_id_to_name/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_index_to_name db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagIndexToNameDb[hash] = &retDB

	return &retDB
}

var TAG_INDEX_LENGTH = 2
var STAT_MAX_TAG_INDEX_PREFIX = []byte (string(config.STAT_KEY_PREX) + "_MAX_TAG_INDEX")
func WriteATag(tag []byte) error {

	tagNameToIndexDB := InitTagNameToIndexDB()

	exsistsIndex := tagNameToIndexDB.ReadFor(tag)
	//has exsited
	if nil != exsistsIndex{
		return nil
	}

	tagIndexToNameDB := InitTagIndexToNameDB()

	maxTagIndex := tagIndexToNameDB.ReadFor(STAT_MAX_TAG_INDEX_PREFIX)
	if 0 == len(maxTagIndex){
		maxTagIndex = []byte{0,0}
	}
	if !fileUtil.BytesIncrement(maxTagIndex){
		return 	errors.New("tag increment max error")
	}
	tagIndexToNameDB.WriteTo(STAT_MAX_TAG_INDEX_PREFIX, maxTagIndex)
	tagIndexToNameDB.WriteTo(maxTagIndex, tag)
	tagNameToIndexDB.WriteTo(tag, maxTagIndex)
	return nil
}


//--------------------------------------------------------------
/**
	tag 库
	格式: tag index --> tag name
 */
var initedTagNameToIndexDb map[int] *DBConfig

func InitTagNameToIndexDB() *DBConfig {
	if nil == initedTagNameToIndexDb {
		initedTagNameToIndexDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagNameToIndexDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/tag_index_to_name/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_name_to_index db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagNameToIndexDb[hash] = &retDB

	return &retDB
}

type TagNameToIndex struct {
	TagName []byte
	TagIndex []byte
}

func GetAllTagNameToIndex() []TagNameToIndex {
	return QueryTagNameToIndex(nil)
}

func QueryTagNameToIndex(tagName []byte) []TagNameToIndex  {
	var ret []TagNameToIndex = nil

	var start []byte = tagName
	var limit []byte = nil
	if nil != tagName{
		limit = fileUtil.CopyBytesTo(tagName)
		fileUtil.BytesIncrement(limit)
	}

	db := InitTagNameToIndexDB()
	iter := db.DBPtr.NewIterator(&util.Range{Start:start, Limit:limit}, &db.ReadOptions)
	iter.First()
	for iter.Valid(){
		if len(iter.Value()) == TAG_INDEX_LENGTH{
			ret = append(ret, TagNameToIndex{TagName: iter.Key(), TagIndex:iter.Value()})
		}
		iter.Next()
	}
	iter.Release()
	return ret
}