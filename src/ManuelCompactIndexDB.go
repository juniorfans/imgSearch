package main

import (
	"fmt"
	"dbOptions"
	"github.com/syndtr/goleveldb/leveldb/util"
	"bufio"
	"os"
	"strings"
	"strconv"
)

func main()  {

	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Println("input dbIds, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")
	for _, dbIdStr := range dbIdStrArray{
		dbIdS,_ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)

		imgToIndexDB := dbOptions.InitMuImgToIndexDb(curDbId)
		indexToImgDB := dbOptions.InitMuIndexToImgDB(curDbId)
		imgToClipDB := dbOptions.InitMuClipToIndexDb(curDbId)
		clipToImgDB := dbOptions.InitMuClipToIndexDb(curDbId)

		imgToIndexDB.DBPtr.CompactRange(util.Range{nil,nil})
		indexToImgDB.DBPtr.CompactRange(util.Range{nil,nil})
		imgToClipDB.DBPtr.CompactRange(util.Range{nil,nil})
		clipToImgDB.DBPtr.CompactRange(util.Range{nil,nil})

		imgToIndexDB.CloseDB()
		indexToImgDB.CloseDB()
		imgToClipDB.CloseDB()
		clipToImgDB.CloseDB()
	}

}