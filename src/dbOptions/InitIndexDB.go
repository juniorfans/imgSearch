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
)

var INDEX_DB_DIR_BASE = "E:/search/"

type DBConfig struct {
	Dir          string
	DBPtr        *leveldb.DB
	OpenOptions  opt.Options
	ReadOptions  opt.ReadOptions
	WriteOptions opt.WriteOptions
	inited       bool

	Name	string
	Id           uint8
	dbType	uint8	//0:source_db 1:index_db

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

/*
var imgClipsReverseIndexDBConfig = &DBConfig{
	Dir : "img_clips_index_reverse/clips.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img clip db",
	dbType:1,
	initParams:nil,
}


var imgIndexToImgDBConfig = &DBConfig{
	Dir : "img_index/img_index.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"index to img db",
	dbType:1,
	initParams:nil,
}



var imgToIndexDBConfig = &DBConfig{
	Dir : "img_to_index/img_to_index.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img to index db",
	dbType:1,
}


var imgClipToIndexDBConfig = &DBConfig{
	Dir : "img_clips_index/clips.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img to clips index db",
	dbType:1,
}


*/

/*
func InitIndexToImgDB() *DBConfig {
	return InitIndexDBByBaseDir(255,3)
}

func InitImgToIndexDB() *DBConfig {
	return InitIndexDBByBaseDir(255,4)
}

func InitIndexToClipDB() *DBConfig {
	return InitIndexDBByBaseDir(255,2)
}

func InitClipToIndexDB() *DBConfig {
	return InitIndexDBByBaseDir(255,1)
}
*/

var initedIndexDb map[int] *DBConfig

func GetTotalMuIndexToClipDB() *DBConfig {
	return InitIndexDBByBaseDir(255,2)
}

func GetTotalMuClipToIndexDb() *DBConfig {
	return InitIndexDBByBaseDir(255,1)
}


func GetTotalMuIndexToImgDB() *DBConfig {
	return InitIndexDBByBaseDir(255,3)
}


func GetTotalMuImgToIndexDb() *DBConfig {
	return InitIndexDBByBaseDir(255,4)
}

func InitMuIndexToClipDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,2)
}

func InitMuClipToIndexDb(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,1)
}


func InitMuIndexToImgDB(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,3)
}


func InitMuImgToIndexDb(dbId uint8) *DBConfig {
	return InitIndexDBByBaseDir(dbId,4)
}


func InitIndexDBByBaseDir(dbId uint8, whichDB int) *DBConfig{

	if nil == initedIndexDb{
		initedIndexDb = make(map[int] *DBConfig)
	}

	hash := (whichDB << 8) + int(dbId)
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

	switch whichDB {
	case 1:
		indexDB.Name = "clip_ident_to_index/data.db"
		break
	case 2:
		indexDB.Name = "clip_index_to_ident/data.db"
		break
	case 3:
		indexDB.Name = "img_index_to_ident/data.db"
		break
	case 4:
		indexDB.Name = "img_ident_to_index/data.db"
		break
	default:
		fmt.Println("whichDB must be 1,2,3,4")
		os.Exit(-1)
		break
	}

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
	iter := InitMuIndexToClipDB(dbId).DBPtr.NewIterator(nil, &opt.ReadOptions{})

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
	removeClosed()
}