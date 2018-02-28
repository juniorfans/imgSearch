package dbOptions

import (
	"imgIndex"
	"fmt"
	"config"
)

var fixClipIndexToIdentSig chan uint8
var fixClipStatIndexToIdentsSig chan uint8

func fixClipIndexToIdent(dbId uint8)  {
	srcDB := InitMuIndexToClipMiddleDB(dbId)
	targetDB := InitMuIndexToClipDB(dbId)
	CombineDBByKeyPrefix(srcDB, targetDB, ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN, 1, config.STAT_KEY_PREX)
	fixClipIndexToIdentSig <- dbId
}

func fixClipStatIndexToIdents(dbid uint8)  {
	srcDB := InitClipStatIndexToIdentsMiddleDB(dbid)
	targetDB := InitClipStatIndexToIdentsDB(dbid)
	CombineDBByKeyPrefix(srcDB, targetDB, ImgIndex.CLIP_STAT_INDEX_BYTES_LEN, 1, config.STAT_KEY_PREX)
	fixClipStatIndexToIdentsSig <- dbid
}

func FixClipIndexToIdentDBs(dbIds []uint8)  {
	fixClipIndexToIdentSig = make(chan uint8, len(dbIds))
	for i,_ := range dbIds{
		go fixClipIndexToIdent(dbIds[i])
	}

	for i:=0;i < len(dbIds);i ++{
		dbId := <-fixClipIndexToIdentSig
		fmt.Println("db ", dbId, " finished~")
	}
	fmt.Println("all finished ~")
}

func FixClipStatIndexToIdentsDBs(dbIds []uint8)  {
	fixClipStatIndexToIdentsSig = make(chan uint8, len(dbIds))
	for i,_ := range dbIds{
		go fixClipStatIndexToIdents(dbIds[i])
	}

	for i:=0;i < len(dbIds);i ++{
		dbId := <-fixClipStatIndexToIdentsSig
		fmt.Println("db ", dbId, " finished~")
	}
	fmt.Println("all finished~")
}

func FixCoordinateIndexDB()  {
	srcDB := InitClipCoordinateIndexToVTagIdMiddleDB()
	targetDB := InitClipCoordinateIndexToVTagIdDB()

	//使用 stat index1 | stat index2 作为键.
	//注意反向的 stat index2 | stat index1 也会在库中, 在前面的 FlushCache 中写入了
	CombineDBByKeyPrefix(srcDB, targetDB, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN, 1, config.STAT_KEY_PREX)
}