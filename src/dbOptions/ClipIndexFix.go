package dbOptions

import (
	"imgIndex"
	"fmt"
)

var fixClipIndexToIdentSig chan uint8
var fixClipStatIndexToIdentsSig chan uint8

func fixClipIndexToIdent(dbId uint8)  {
	srcDB := InitMuIndexToClipMiddleDB(dbId)
	targetDB := InitMuIndexToClipDB(dbId)
	CombineDBByKeyPrefix(srcDB, targetDB, ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN, true)
	fixClipIndexToIdentSig <- dbId
}

func fixClipStatIndexToIdents(dbid uint8)  {
	srcDB := InitClipStatIndexToIdentsMiddleDB(dbid)
	targetDB := InitClipStatIndexToIdentsDB(dbid)
	CombineDBByKeyPrefix(srcDB, targetDB, ImgIndex.CLIP_STAT_INDEX_BYTES_LEN, true)
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