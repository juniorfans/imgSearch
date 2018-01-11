package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"errors"
	"strconv"
)

type ImgDBConfig struct {
	img_dir      string
	DBPtr        *leveldb.DB
	OpenOptions  opt.Options
	ReadOptions  opt.ReadOptions
	WriteOptions opt.WriteOptions
	inited bool
}


var imgDBConfig ImgDBConfig

func PickImgDB(index int) *ImgDBConfig {
	dbDir := "D:/img_db_" +  strconv.Itoa(index)+ "/image.db"

	fmt.Println("has pick this img db: ", dbDir)

	imgDBConfig = ImgDBConfig{
		img_dir : dbDir,
		DBPtr : nil,
		OpenOptions : opt.Options{ErrorIfMissing:false},
		ReadOptions : opt.ReadOptions{},
		WriteOptions : opt.WriteOptions{Sync:false},
		inited : false,
	}
	return InitImgDB()
}

var imgClipsIndexDBConfig = ImgDBConfig{
	img_dir : "D:/img_clip_db/clips.db",
	DBPtr : nil,
	OpenOptions : opt.Options{ErrorIfMissing:false},
	ReadOptions : opt.ReadOptions{},
	WriteOptions : opt.WriteOptions{Sync:false},
	inited : false,
}


var imgIndexDBConfig = ImgDBConfig{
	img_dir : "D:/img_index/img_index.db",
	DBPtr : nil,
	OpenOptions : opt.Options{ErrorIfMissing:false},
	ReadOptions : opt.ReadOptions{},
	WriteOptions : opt.WriteOptions{Sync:false},
	inited : false,
}


func InitImgDB() *ImgDBConfig {
	_, err :=  InitDB(&imgDBConfig)
	if err != nil{
		fmt.Println("open img db error, ", err)
		return nil
	}
	return &imgDBConfig
}

func InitImgClipsDB() *ImgDBConfig {
	_, err :=  InitDB(&imgClipsIndexDBConfig)
	if err != nil{
		fmt.Println("open img clip db error, ", err)
		return nil
	}
	return &imgClipsIndexDBConfig
}

func InitImgIndexDB() *ImgDBConfig {
	_, err :=  InitDB(&imgIndexDBConfig)
	if err != nil{
		fmt.Println("open img index db error, ", err)
		return nil
	}
	return &imgIndexDBConfig
}

func InitDB(config *ImgDBConfig) (dbPtr *leveldb.DB, err error) {
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

	config.DBPtr,err = leveldb.OpenFile(config.img_dir, &config.OpenOptions)
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

func ReadClipValues(count int)  {
	InitImgClipsDB()
	iter := imgClipsIndexDBConfig.DBPtr.NewIterator(nil, &opt.ReadOptions{})

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

func (this *ImgDBConfig) CloseDB()  {
	this.DBPtr.Close()
}