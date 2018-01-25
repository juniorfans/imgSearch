package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"config"
	"os"
	"bufio"
	"strings"
	"io"
	"util"
	"bytes"
)

var SECOND_DB_DIR_BASE = "E:/search/"

type DBConfig struct {
	Dir          string
	DBPtr        *leveldb.DB
	OpenOptions  opt.Options
	ReadOptions  opt.ReadOptions
	WriteOptions opt.WriteOptions
	inited       bool

	Name	string
	Id           uint8
	dbType	uint8	//0:source_db 1:second_db
}

var HAS_READ_DB_CONF = false
func readDBDirConf() {
	if HAS_READ_DB_CONF == true{
		return
	}
	HAS_READ_DB_CONF = true
	exedir,err := fileUtil.GetCurrentMoudlePath()
	if nil != err{
		fmt.Println("get current moudle error: ", err)
		return
	}
	f, err := os.Open(exedir + "/" + "config.txt")
	if err != nil {
		fmt.Println("open config error: ", err)
		return
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')

		if io.EOF == err{
			return
		}else if nil != err {
			fmt.Println("read config error")
			return
		}
		line = strings.TrimSpace(line)
		if 0 == len(line){
			continue
		}
		kv := strings.Split(line,"=")
		if 2 == len(kv){
			if 0==strings.Compare("second_db_dir_base",kv[0]){
				SECOND_DB_DIR_BASE = strings.TrimSpace(kv[1])
				fmt.Println("second_db_dir_base: ", SECOND_DB_DIR_BASE)
			}else if 0==strings.Compare("soruce_db_dir_base",kv[0]){
				DB_DIR_BASE = strings.TrimSpace(kv[1])
				fmt.Println("source_db_dir_base: ", DB_DIR_BASE)
			}
		}
	}
	return
}


var imgClipsReverseIndexDBConfig = DBConfig{
	Dir : "img_clips_index_reverse/clips.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img clip db",
	dbType:1,
}


var imgIndexToImgDBConfig = DBConfig{
	Dir : "img_index/img_index.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"index to img db",
	dbType:1,
}

var imgLetterDBConfig = DBConfig{
	Dir : "img_letter/img_letter.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img letter db",
	dbType:1,
}

var imgToIndexDBConfig = DBConfig{
	Dir : "img_to_index/img_to_index.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img to index db",
	dbType:1,
}


var imgClipsIndexDBConfig = DBConfig{
	Dir : "img_clips_index/clips.db",
	DBPtr : nil,
	inited : false,

	Id:0,
	Name:"img to clips index db",
	dbType:1,
}

/**
	 key 	: clip 索引值
	 value	: clip 集合{某个库的某个 mainImgId 的第 which 张子图}
 */
func InitIndexToClipDB() *DBConfig {
	_, err :=  initDB(&imgClipsReverseIndexDBConfig)
	if err != nil{
		fmt.Println("open img clip reverse index db error, ", err)
		return nil
	}
	return &imgClipsReverseIndexDBConfig
}

func InitIndexToImgDB() *DBConfig {
	_, err :=  initDB(&imgIndexToImgDBConfig)
	if err != nil{
		fmt.Println("open img index db error, ", err)
		return nil
	}
	return &imgIndexToImgDBConfig
}

func InitImgLetterDB() *DBConfig {
	_, err :=  initDB(&imgLetterDBConfig)
	if err != nil{
		fmt.Println("open img letter db error, ", err)
		return nil
	}
	return &imgLetterDBConfig
}

func InitImgToIndexDB() *DBConfig {
	_, err :=  initDB(&imgToIndexDBConfig)
	if err != nil{
		fmt.Println("open img to index db error, ", err)
		return nil
	}
	return &imgToIndexDBConfig
}

/**
	key	: clip 信息(某个库的某个 mainImgId 的第 which 张子图)
	value	: 该 clip 的索引
 */
func InitClipsIndexDB() *DBConfig {
	_, err :=  initDB(&imgClipsIndexDBConfig)
	if err != nil{
		fmt.Println("open img to clips index db error, ", err)
		return nil
	}
	return &imgClipsIndexDBConfig
}

func initDB(config *DBConfig) (dbPtr *leveldb.DB, err error) {
	if nil == config{
		dbPtr = nil
		err = errors.New("db config is nil")
		return
	}
	if config.inited{
		dbPtr = config.DBPtr
		err = nil
		return
	}
	{
		config.OpenOptions = opt.Options{
			ErrorIfMissing:false,
			BlockSize:40 * opt.KiB,
			CompactionTableSize:20*opt.MiB,
			BlockCacheCapacity:128 * opt.MiB,
			WriteBuffer:32*opt.MiB,
			CompactionL0Trigger:4,
			CompactionTotalSize:40*opt.MiB,
		}
	}

	{
		config.ReadOptions = opt.ReadOptions{}
	}
	{
		config.WriteOptions = opt.WriteOptions{Sync:false}
	}

	{
		readDBDirConf()
		if uint8(0) == config.dbType{
			config.Dir = DB_DIR_BASE + config.Dir
		}else if uint8(1) == config.dbType{
			config.Dir = SECOND_DB_DIR_BASE + config.Dir
		}

		fmt.Println("has pick this img db: ", config.Dir)
	}

	config.DBPtr,err = leveldb.OpenFile(config.Dir, &config.OpenOptions)
	if err != nil{
		fmt.Println("open db failed")
		return
	}

	config.inited = true

	return
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

func ReadClipValuesInCount(count int)  {
	iter := InitIndexToClipDB().DBPtr.NewIterator(nil, &opt.ReadOptions{})

	if(!iter.First()){
		fmt.Println("seek to first error")
	}

	for iter.Valid(){
		//writeToFile(iter.Value(), string(iter.Key()))
		if -1 != bytes.Index(iter.Key(), config.STAT_KEY_PREX){
			continue
		}

		valueList := ParseImgClipIdentListBytesToStrings(iter.Value())
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