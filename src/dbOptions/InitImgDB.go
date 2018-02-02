package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"strconv"
	"github.com/syndtr/goleveldb/leveldb"
	"errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
)

var imgDBs map[uint8]*DBConfig = make(map[uint8]*DBConfig)

func GetImgDBs() []*DBConfig {
	fmt.Println("picked img db : ", len(imgDBs))
	ret := make([]*DBConfig, len(imgDBs))
	i := 0
	for _,db := range imgDBs{
		ret[i] = db
		i ++
	}
	return ret
}

func PickImgDB(dbId uint8) *DBConfig {
	ret := imgDBs[dbId]
	if nil == ret{
		dbDir := "img_db_" +  strconv.Itoa(int(dbId))+ "/image.db"
		imgDBConfig := DBConfig{
			Dir : dbDir,
			DBPtr : nil,
			OpenOptions : opt.Options{
				ErrorIfMissing:false,
				BlockSize:40 * opt.KiB,
				CompactionTableSize:20*opt.MiB,
				BlockCacheCapacity:64 * opt.MiB,
				Filter:filter.NewBloomFilter(10),
			},
			ReadOptions : opt.ReadOptions{},
			WriteOptions : opt.WriteOptions{Sync:false},
			inited : false,
			Id : dbId,
			Name:"img db",
			dbType:0,	//source db
			initParams:nil,
		}

		_, err :=  initDB(&imgDBConfig)
		if err != nil{
			fmt.Println("open img db error, ", err)
			return nil
		}
		imgDBs[dbId] = &imgDBConfig
		ret = &imgDBConfig
	}else{
		initDB(ret)	//防止 DBConfig 被关闭但没有移除掉, 现在需要复用，所以要重新初始化
	}

	return ret
}

func GetImgDBWhichPicked() *DBConfig {
	if 0 == len(imgDBs){
		fmt.Println("no picked img db")
		return nil
	}else if 1 == len(imgDBs){
		for _,ret := range imgDBs{
			return ret
		}
		fmt.Println("this must not be happen!")
		return nil
	}else{
		fmt.Println("too many picked img db")
		return nil
	}
}

func removeClosed()  {
	var aliveDBs []*DBConfig
	for _,db := range imgDBs{
		if false != db.inited{
			aliveDBs = append(aliveDBs,db)
		}
	}

	imgDBs = make(map[uint8]*DBConfig, len(aliveDBs))
	for _,db := range aliveDBs{
		imgDBs[db.Id]=db
	}
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

	if nil == config.initParams{
		if 0 == config.dbType{
			config.initParams = ReadDBConf("conf_img_db.txt")
		}else if 1 == config.dbType{
			fmt.Println("error, init index db must use InitIndexDB")
			dbPtr = nil
			err = errors.New("init index db must use InitIndexDB")
		}

		config.Dir = config.initParams.DirBase + config.Dir
		config.OpenOptions = *getLevelDBOpenOption(config.initParams)
		fmt.Println("has pick this img db: ", config.Dir)
		config.initParams.PrintLn()
	}

	{
		config.ReadOptions = opt.ReadOptions{}
	}
	{
		config.WriteOptions = opt.WriteOptions{Sync:false}
	}
	config.DBPtr,err = leveldb.OpenFile(config.Dir, &config.OpenOptions)
	if err != nil{
		fmt.Println("open db failed")
		return
	}

	config.inited = true

	return
}