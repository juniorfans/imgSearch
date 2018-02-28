package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

//每处理一张大图, 会分配一个 virtual tag id

//clip coordinate index 库
//key: statIndex1 | statIndex2
//value: {clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2 | vtag | support}, 上面结构可能是重复的
func InitClipCoordinateIndexToVTagIdDB() *DBConfig {
	return innerInitClipCoordinateIndexToVTagIdDB(false)
}

//clip coordinate index 的中间库
//key: statIndex1 | statIndex2 | clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2 | vtag | support
//value: nil
func InitClipCoordinateIndexToVTagIdMiddleDB() *DBConfig {
	return innerInitClipCoordinateIndexToVTagIdDB(true)
}

var initedClipCoordinateBranchIndexToVirtualDb map[int] *DBConfig
func innerInitClipCoordinateIndexToVTagIdDB(isMiddle bool) *DBConfig {
	if nil == initedClipCoordinateBranchIndexToVirtualDb {
		initedClipCoordinateBranchIndexToVirtualDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)

	if isMiddle{
		hash = hash << 8 + 1
	}

	if exsitsDB, ok := initedClipCoordinateBranchIndexToVirtualDb[hash];ok && true == exsitsDB.inited{
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

	if isMiddle{
		retDB.Name = "result/clip_coordinate_bindex_vtag_middle/data.db"
	}else{
		retDB.Name = "result/clip_coordinate_bindex_vtag/data.db"
	}

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_coordinate_bindex_vtag db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipCoordinateBranchIndexToVirtualDb[hash] = &retDB

	return &retDB
}

/**
	coordinate index 库的反向索引库
	key: vtag | clipIndex1 | clipIndex2
	value: support
	此库主要用于测试: 验证 coordinate 关系是否真实
 */
var initedClipCoordinateVTagIdToBranchIndexDb map[int] *DBConfig
func InitClipCoordinatevTagIdToIndexDB() *DBConfig {
	if nil == initedClipCoordinateVTagIdToBranchIndexDb {
		initedClipCoordinateVTagIdToBranchIndexDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipCoordinateVTagIdToBranchIndexDb[hash];ok && true == exsitsDB.inited{
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

	retDB.Name = "result/clip_coordinate_vtag_clipident/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_coordinate_vtag_clipident db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipCoordinateVTagIdToBranchIndexDb[hash] = &retDB

	return &retDB
}

var initedNotClipCoordinateIndexDB map[int] *DBConfig
func InitNotClipCoordinateIndexDB() *DBConfig {
	if nil == initedNotClipCoordinateIndexDB {
		initedNotClipCoordinateIndexDB = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedNotClipCoordinateIndexDB[hash];ok && true == exsitsDB.inited{
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

	retDB.Name = "result/clip_not_coordinate_index/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_not_coordinate_index db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedNotClipCoordinateIndexDB[hash] = &retDB

	return &retDB
}

