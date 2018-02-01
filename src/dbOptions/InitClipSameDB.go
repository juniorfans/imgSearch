package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strconv"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

var initedClipSameDb map[int] *DBConfig

func InitClipSameDB(dbId uint8) *DBConfig {
	if nil == initedIndexDb{
		initedIndexDb = make(map[int] *DBConfig)
	}
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
		dbType:1,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_index_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "clip_same/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/" + strconv.Itoa(int(dbId)) + "/" + retDB.Name
	fmt.Println("has pick this clip same db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipSameDb[hash] = &retDB

	return &retDB
}