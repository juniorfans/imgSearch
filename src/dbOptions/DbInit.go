package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"errors"
)

type DBConfig struct {
	Dir          string
	DBPtr        *leveldb.DB
	OpenOptions  opt.Options
	ReadOptions  opt.ReadOptions
	WriteOptions opt.WriteOptions
	inited       bool

	Id           uint8
}


var imgClipsIndexDBConfig = DBConfig{
	Dir : "D:/img_clip_db/clips.db",
	DBPtr : nil,
	OpenOptions : opt.Options{ErrorIfMissing:false},
	ReadOptions : opt.ReadOptions{},
	WriteOptions : opt.WriteOptions{Sync:false},
	inited : false,
}


var imgIndexDBConfig = DBConfig{
	Dir : "D:/img_index/img_index.db",
	DBPtr : nil,
	OpenOptions : opt.Options{ErrorIfMissing:false},
	ReadOptions : opt.ReadOptions{},
	WriteOptions : opt.WriteOptions{Sync:false},
	inited : false,
}

var imgLetterDBConfig = DBConfig{
	Dir : "D:/img_letter/img_letter.db",
	DBPtr : nil,
	OpenOptions : opt.Options{ErrorIfMissing:false},
	ReadOptions : opt.ReadOptions{},
	WriteOptions : opt.WriteOptions{Sync:false},
	inited : false,
}


func InitImgClipsDB() *DBConfig {
	_, err :=  initDB(&imgClipsIndexDBConfig)
	if err != nil{
		fmt.Println("open img clip db error, ", err)
		return nil
	}
	return &imgClipsIndexDBConfig
}

func InitImgIndexDB() *DBConfig {
	_, err :=  initDB(&imgIndexDBConfig)
	if err != nil{
		fmt.Println("open img index db error, ", err)
		return nil
	}
	return &imgIndexDBConfig
}

func InitImgLetterDB() *DBConfig {
	_, err :=  initDB(&imgLetterDBConfig)
	if err != nil{
		fmt.Println("open img letter db error, ", err)
		return nil
	}
	return &imgLetterDBConfig
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

	config.DBPtr,err = leveldb.OpenFile(config.Dir, &config.OpenOptions)
	if err != nil{
		fmt.Println("open db failed")
		return
	}

	config.inited = true

	return
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
	iter := imgClipsIndexDBConfig.DBPtr.NewIterator(nil, &opt.ReadOptions{})

	if(!iter.First()){
		fmt.Println("seek to first error")
	}

	for iter.Valid(){
		//writeToFile(iter.Value(), string(iter.Key()))
		valueList := ParseClipIndeValues(iter.Value())
		valueList.Print()
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
}