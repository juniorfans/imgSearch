package main

import (
	"os"
	"fmt"
	"bufio"
	"strings"
	"strconv"
	"dbOptions"
	"config"
)

func main()  {

	clipConfig := config.GetClipConfigById(0)

	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Println("input dbIds, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")
	for _, dbIdStr := range dbIdStrArray{
		dbIdS,_ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)

		clipIdentToIndexDB := dbOptions.InitIndexDBByBaseDir(curDbId, 1)
		clipIndexToIdentDB := dbOptions.InitIndexDBByBaseDir(curDbId, 2)
		imgDB := dbOptions.PickImgDB(curDbId)

		fmt.Println("begin to deal imgdb: ", strconv.Itoa(int(curDbId)))
		dbOptions.BeginImgClipSaveEx(curDbId,-1, clipConfig.ClipOffsets , clipConfig.ClipLengh)
		fmt.Println("end of deal imgdb: ", strconv.Itoa(int(curDbId)))
		clipIdentToIndexDB.CloseDB()
		clipIndexToIdentDB.CloseDB()
		imgDB.CloseDB()
	}

}
