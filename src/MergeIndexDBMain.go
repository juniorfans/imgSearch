package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"strconv"
	"dbOptions"
)

func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Println("input dbIds, to merge index db, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")

	//clipToIndexMergeDB := dbOptions.GetTotalMuClipToIndexDb()
	//imgToIndexMergeDB := dbOptions.GetTotalMuImgToIndexDb()

	indexToClipMergeDB := dbOptions.GetTotalMuIndexToClipDB()

	indexToImgMergeDB := dbOptions.GetTotalMuIndexToImgDB()

	for _, dbIdStr := range dbIdStrArray {
		dbIdS, _ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)
		/*
			clipToIndexDB := dbOptions.InitMuClipToIndexDb(curDbId)
			fmt.Println("begin to merge ", clipToIndexDB.Name)
			dbOptions.MergeTo(clipToIndexDB, clipToIndexMergeDB)
		*/
		indexToClipDB := dbOptions.InitMuIndexToClipDB(curDbId)
		fmt.Println("begin to merge ", indexToClipDB.Name)
		dbOptions.MergeTo(indexToClipDB, indexToClipMergeDB)

		/*
			imgToIndexDB := dbOptions.InitMuImgToIndexDb(curDbId)
			fmt.Println("begin to merge ", imgToIndexDB.Name)
			dbOptions.MergeTo(imgToIndexDB, imgToIndexMergeDB)
		*/

		indexToImgDB := dbOptions.InitMuIndexToImgDB(curDbId)
		fmt.Println("begin to merge ", indexToImgDB.Name)
		dbOptions.MergeTo(indexToImgDB, indexToImgMergeDB)
	}
}
