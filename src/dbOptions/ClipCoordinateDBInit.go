package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

//每处理一张大图, 会分配一个 virtual tag id

/**
库名称:	clip 到虚拟 tag 映射库, 用于判断两个子图是否具备协同关系
库名: 	coordinate_clip_to_vtag
初始化:	InitCoordinateClipToVTagMiddleDB
格式:	InitCoordinateClipToVTagDB: statIndex1 | statIndex2 --> clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2 | vtag | support
说明:	值可重复
 */
func InitCoordinateClipToVTagDB() *DBConfig {
	return innerInitClipCoordinateIndexToVTagIdDB(false)
}

/**
库名称:	clip 到虚拟 tag 映射库中间库
库名:	coordinate_clip_to_vtag_middle
初始化:	InitCoordinateClipToVTagMiddleDB
格式:	statIndex1 |  statIndex2 | clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2 | vtag | support --> nil
 */
func InitCoordinateClipToVTagMiddleDB() *DBConfig {
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
		retDB.Name = "result/coordinate_clip_to_vtag_middle/data.db"
	}else{
		retDB.Name = "result/coordinate_clip_to_vtag/data.db"
	}

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this ", retDB.Name ," db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipCoordinateBranchIndexToVirtualDb[hash] = &retDB

	return &retDB
}

/**
库名称:	虚拟 tag 到 clip 映射库. 用于测试: 验证 coordinate 关系是否真实
库名:	coordinate_vtag_to_clip
初始化:	InitCoordinatevTagToClipDB
格式:	vtag | clipIndex1 | clipIndex2 -> suppot
说明:	值可重复
 */
var initedCoordinateVTagToClipDb map[int] *DBConfig
func InitCoordinatevTagToClipDB() *DBConfig {
	if nil == initedCoordinateVTagToClipDb {
		initedCoordinateVTagToClipDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedCoordinateVTagToClipDb[hash];ok && true == exsitsDB.inited{
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

	retDB.Name = "result/coordinate_vtag_to_clip/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this coordinate_vtag_to_clip db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedCoordinateVTagToClipDb[hash] = &retDB

	return &retDB
}

var initedNotSameTopicDB map[int] *DBConfig
func InitNotSameTopicDB() *DBConfig {
	if nil == initedNotSameTopicDB {
		initedNotSameTopicDB = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedNotSameTopicDB[hash];ok && true == exsitsDB.inited{
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

	retDB.Name = "result/not_same_topic/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this not_same_topic db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedNotSameTopicDB[hash] = &retDB

	return &retDB
}

