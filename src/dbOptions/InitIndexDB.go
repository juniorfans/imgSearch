package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"config"
	"bytes"
	"os"
	"strconv"
	"imgIndex"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"util"
)

/**
---------------------------------------------------- 一些名词的解释 ---------------------------
虚拟 tag (virtual tag, vtag): 每处理一张大图都是一个独立的 tag, 协同关系关联一个 vtag 表明协同关系局限在这张大图里面.
暂时只有一个地方被使用: GetCoordinateSupport

值可重复: 指的是库的值是重复的结构, 可以包括多个重复单元. 如 statIndexToClip 表, 相同的 stat index 对应着多个 clipIdent, 将它们顺序叠加作为值.


 */

/**
---------------------------------------------------- img 索引库---------------------------

-----------------------------
库名称:	img 到索引映射库
库名:	img_ident_to_index
初始化:	InitImgToIndexDB
格式:	imgIdent --> imgIndex
用途:	用于快速找出图片答案, 排查图片是否已经处理过, 协同计算时计算共同出现在一张大图中的子图集合(使用img索引作为键)

-----------------------------
库名称:	img 索引到 img 映射库
库名:	img_index_to_index
初始化:	InitIndexToImgDB
格式: 	imgIndex --> imgIdents
用途:	目前主要用于测试

---------------------------------------------------- clip 索引库---------------------------
库名称:	clip 分支索引映射 ident 库
库名:	clip_index_to_ident
初始化:	InitIndexToClipDB
格式:	clipBranchIndex --> clipIdents
说明:	值可重复
用途:	暂时只用于测试

-----------------------------
库名称:	clip 分支索引映射 ident 库中间库
库名:	clip_index_to_ident_middle
初始化:	InitIndexToClipMiddleDB
格式:	clipBranchIndex | clipIdent --> nil
用途:	分裂键, 得到的两部分, 前者作为 clip_index_to_ident 表的键, 后者作为值

-----------------------------
库名称:	clip 统计信息索引到 ident 库
库名:	clip_stat_index_to_ident
初始化:	InitStatIndexToClipDB
格式:	clipStatIndex --> clipIdents
说明:	值可重复
用途:	协同分析时, 使用目标子图的 stat index 查找该库, 得出相同 stat index 的有哪些 clipIdent 集合, 再得出 imgIdent 集合, 即是目标子图出现的母图. 之后再计算多个目标子图的重合.
	注意计算重合时不可用 imgIdent 为键, 而应使用 imgIndex 为鍵

-----------------------------
库名称:	clip 统计信息索引到 ident 库中间库
库名:	clip_stat_index_to_ident_middle
初始化:	InitStatIndexToClipMiddleDB
格式:	clipStatIndex | clipIdent --> nil
用途:	分裂键得到的键值插入 clip_stat_index_to_ident

-----------------------------
库名称:	clip ident 到原索引库
库名:	clip_ident_to_index
初始化:	InitClipToIndexDB
格式:	clip ident --> clip source index
说明:	注意是原索引, 不是分支索引或统计索引
用途:	用于协同计算时给定大图的 imgIdent 和 子图编号 whiches, 查找得到 clip index, 进一步得到 stat index


---------------------------------------------------- 协同库---------------------------
库名称:	clip 到虚拟 tag 映射库
库名: 	coordinate_clip_to_vtag
初始化:	InitCoordinateClipToVTagDB
格式:	statIndex1 | statIndex2 --> clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2 | vtag | support
说明:	值可重复
用途:	计算出协同关系后, 保存结果; 用于后续查找两个子图是否具有协同关系

-----------------------------
库名称:	clip 到虚拟 tag 映射库中间库
库名:	coordinate_clip_to_vtag_middle
初始化:	InitCoordinateClipToVTagMiddleDB
格式:	statIndex1 |  statIndex2 | clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2 | vtag | support --> nil
用途:	分裂键, 得到键值插入 coordinate_clip_to_vtag

-----------------------------
库名称:	虚拟 tag 到 clip 映射库.
库名:	coordinate_vtag_to_clip
初始化:	InitCoordinatevTagToClipDB
格式:	vtag | clipIndex1 | clipIndex2 -> suppot
说明:	值可重复
用途:	用于测试: 验证 coordinate 关系是否真实

-----------------------------
库名称:	非主题相似子图库
库名:	not_same_topic
初始化:	InitNotSameTopicDB
格式:	statIndex1 |  statIndex2 --> clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2
说明:	值可重复
用途:	用于进一步判断具有协同关系的两个子图是否主题相似


---------------------------------------------------- 训练库---------------------------
训练库的格式:
库名称:	子图-标签库
库名:	clip_to_tag
初始化:	InitClipToTagDB
格式:	clipStatIndex -> clipIndex | clipIdent | tagId
说明:	值可重复
用途:	用于接收训练时给子图打的标签; 识别图片时找出大图中哪些子图具有相同的标签

-----------------------------
库名称:	标签-子图库
库名:	tag_to_clip
初始化:	InitTagToClipDB
格式:	tagId -> statInex|clipIndex|clipIdent
说明:	值可重复
用途:	目前只用于测试; 以后规划: 识别出图片中的文字与标签比对直接得出大图中有哪些子图是此标签

-----------------------------
库名称:	大图-答案库
库名:	img_answer
初始化:	InitImgAnswerDB
格式:	imgIndexBytes --> which array
说明:	库的值直接就是多个字节, 表示此大图的这些子图是答案
用途:	用于立即定位大图的答案

-----------------------------
库名称:	标签id 到名称的映射
库名:	tag_id_to_name
初始化:	InitTagIdToNameDB
格式:	tagId --> tagName
用途:	训练时保存录入的主题

-----------------------------
库名称:	标签名称到 id 的映射
库名:	tag_name_to_id
初始化:	InitTagNameToIdDB
格式:	tagName --> tagId
用途:	训练时保存录入的主题
 */

type DBConfig struct {
	Dir          string
	DBPtr        *leveldb.DB
	OpenOptions  opt.Options
	ReadOptions  opt.ReadOptions
	WriteOptions opt.WriteOptions
	inited       bool

	Name	string
	Id           uint8
	dbType	uint8	//0:source_db 1:index_db 2:result_db

	initParams *DBInitParams
}

var imgLetterDBConfig = DBConfig{
	Dir : "img_letter/img_letter.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img letter db",
	dbType:1,
}

func InitImgLetterDB() *DBConfig {
	_, err :=  initDB(&imgLetterDBConfig)
	if err != nil{
		fmt.Println("open img letter db error, ", err)
		return nil
	}
	return &imgLetterDBConfig
}

var initedIndexDb map[int] *DBConfig

//clip index to ident. 键是分支索引
func InitIndexToClipDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,2, false)
}
//中间表
func InitIndexToClipMiddleDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,2, true)
}

//clip ident to source clip index. 不是分支索引，也不含统计字节
func InitClipToIndexDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,1, false)
}


func InitIndexToImgDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,3, false)
}


func InitImgToIndexDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,4, false)
}

func GetInitedClipIdentToIndexDB() []*DBConfig {

	var ret []*DBConfig
	for hash,db :=range initedIndexDb{
		whichDB := hash >> 8
		if 1 == whichDB{
			ret = append(ret, db)
		}
	}
	return ret
}

func GetInitedImgIndexToIdentDB() []*DBConfig {

	var ret []*DBConfig
	for hash,db :=range initedIndexDb{
		whichDB := hash >> 8
		if 3 == whichDB{
			ret = append(ret, db)
		}
	}
	return ret
}

func GetInitedImgIdentToIndexDB() []*DBConfig {

	var ret []*DBConfig
	for hash,db :=range initedIndexDb{
		whichDB := hash >> 8
		if 4 == whichDB{
			ret = append(ret, db)
		}
	}
	return ret
}

func InitIndexDBByBaseDir(dbId uint8, whichDB int, isMiddle bool) *DBConfig{

	if nil == initedIndexDb{
		initedIndexDb = make(map[int] *DBConfig)
	}

	hash := (whichDB << 8) + int(dbId)
	if isMiddle{
		hash = hash << 8 + 1
	}
	if exsitsDB, ok := initedIndexDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	indexDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:1,
		}

	if nil == indexDB.initParams{
		indexDB.initParams = ReadDBConf("conf_index_db.txt")
		indexDB.OpenOptions = *getLevelDBOpenOption(indexDB.initParams)
		indexDB.initParams.PrintLn()
	}

	{
		indexDB.ReadOptions = opt.ReadOptions{}
	}
	{
		indexDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	indexDB.inited = true
	var dbName string
	switch whichDB {
	case 1:
		dbName = "clip_ident_to_index"
		break
	case 2:
		dbName = "clip_index_to_ident"
		break
	case 3:
		dbName = "img_index_to_ident"
		break
	case 4:
		dbName = "img_ident_to_index"
		break
	default:
		fmt.Println("whichDB must be 1,2,3,4")
		os.Exit(-1)
		break
	}

	if isMiddle{
		dbName += "_middle"
	}
	indexDB.Name = dbName + "/data.db"
	indexDB.Dir = indexDB.initParams.DirBase + "/" + strconv.Itoa(int(dbId)) + "/" + indexDB.Name
	fmt.Println("has pick this index db: ", indexDB.Dir)
	indexDB.DBPtr,_ = leveldb.OpenFile(indexDB.Dir, &indexDB.OpenOptions)
	indexDB.inited = true

	initedIndexDb[hash] = &indexDB

	return &indexDB
}


func getLevelDBOpenOption(dbParams *DBInitParams) *opt.Options {
	return &opt.Options{
		ErrorIfMissing:false,
		BlockSize: dbParams.BlockSize,
		CompactionTableSize: dbParams.CompactionTableSize,
		BlockCacheCapacity: dbParams.BlockCacheCapacity,
		WriteBuffer: dbParams.WriteBuffer,
		CompactionL0Trigger: dbParams.CompactionL0Trigger,
		CompactionTotalSize: dbParams.CompactionTotalSize,
		Filter:filter.NewBloomFilter(10),
	}
}

func (this *DBConfig)WriteTo(key , value[]byte) error {
	return this.DBPtr.Put(key, value, &this.WriteOptions)
}


func (this *DBConfig)WriteBatchTo (batch *leveldb.Batch) {
	this.DBPtr.Write(batch, &this.WriteOptions)
}

func (this *DBConfig)ReadFor(key []byte) []byte {
	ret, err := this.DBPtr.Get(key,&this.ReadOptions)
	if err == leveldb.ErrNotFound{
		return nil
	}
	return ret
}

func (this *DBConfig) IsEmpty() bool {
	iter := this.DBPtr.NewIterator(nil, &this.ReadOptions)
	defer iter.Release()

	if false == iter.First(){
		return true
	}
	for iter.Valid(){
		if !fileUtil.BytesStartWith(iter.Key(), config.STAT_KEY_PREX){
			return false
		}
		iter.Next()
	}

	return true
}

func (this *DBConfig) PrintStat()  {
	limit := make([]byte, len(config.STAT_KEY_PREX)+10)
	ci := 0
	ci += copy(limit[ci:], config.STAT_KEY_PREX)
	ci += copy(limit[ci:], []byte{255,255,255,255,255,255,255,255,255,255})
	region := util.Range{Start:config.STAT_KEY_PREX, Limit:limit}
	iter := this.DBPtr.NewIterator(&region, &this.ReadOptions)
	iter.First()

	fmt.Println("----------------------- begin ----------------------------")
	fmt.Println("dbname: ", this.Name, ", id: ", this.Id)
	for iter.Valid(){
		fmt.Println(string(iter.Key()),  " : ", string(iter.Value()))
		iter.Next()
	}

	fmt.Println("----------------------- end ----------------------------")
}

func ReadKeys(dbPtr *leveldb.DB, count int)  {
	iter := dbPtr.NewIterator(nil, &opt.ReadOptions{})

	if(!iter.First()){
		fmt.Println("seek to first error")
	}

	for iter.Valid(){
		//writeToFile(iter.Value(), string(iter.Key()))
		fmt.Println((iter.Key()))
		iter.Next()
		count --
		if count <= 0{
			break
		}
	}
	iter.First()
}

func ReadClipValuesInCount(dbId uint8, count int)  {
	iter := InitIndexToClipDB(dbId).DBPtr.NewIterator(nil, &opt.ReadOptions{})

	if(!iter.First()){
		fmt.Println("seek to first error")
	}

	for iter.Valid(){
		//writeToFile(iter.Value(), string(iter.Key()))
		if -1 != bytes.Index(iter.Key(), config.STAT_KEY_PREX){
			continue
		}

		valueList := ImgIndex.FromClipIdentsToStrings(iter.Value())
	//	for _, valueStr := range valueList{
	//		fmt.Println(valueStr)
	//	}
		fmt.Println(valueList)
		iter.Next()
		count --
		if count <= 0{
			break
		}
	}
	iter.First()
}

func ReadValues(dbPtr *leveldb.DB, count int)  {
	iter := dbPtr.NewIterator(nil, &opt.ReadOptions{})

	if(!iter.First()){
		fmt.Println("seek to first error")
	}

	for iter.Valid(){
		//writeToFile(iter.Value(), string(iter.Key()))
		fmt.Println(string(iter.Value()))
		iter.Next()
		count --
		if count <= 0{
			break
		}
	}
	iter.First()
}

func (this *DBConfig) CloseDB()  {
	this.inited = false
	this.DBPtr.Close()

	if 0 == this.dbType{
		markImgDBClosed()
	}

}