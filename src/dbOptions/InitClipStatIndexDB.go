package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
)


func InitClipStatIndexToIdentsDB(dbId uint8) *DBConfig {
	return innerInitClipStatIndexDB(dbId, false)
}

func InitClipStatIndexToIdentsMiddleDB(dbId uint8) *DBConfig {
	return innerInitClipStatIndexDB(dbId, true)
}


var initedClipStatIndexDB map[int] *DBConfig
func innerInitClipStatIndexDB(dbId uint8, isMiddle bool) *DBConfig {
	if nil == initedClipStatIndexDB {
		initedClipStatIndexDB = make(map[int] *DBConfig)
	}

	hash := int(dbId)
	if isMiddle{
		hash = hash << 8 + 1
	}
	if exsitsDB, ok := initedClipStatIndexDB[hash];ok && true == exsitsDB.inited{
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
		retDB.initParams = ReadDBConf("conf_index_db.txt")
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

	if isMiddle{
		retDB.Name = "clip_stat_index_to_ident_middle/data.db"
	}else{
		retDB.Name = "clip_stat_index_to_ident/data.db"
	}


	retDB.Dir = retDB.initParams.DirBase + "/" + strconv.Itoa(int(dbId)) + "/" + retDB.Name
	fmt.Println("has pick this index db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipStatIndexDB[hash] = &retDB

	return &retDB
}
