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
	fmt.Println("input dbIds, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")
	for _, dbIdStr := range dbIdStrArray{
		dbIdS,_ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)

		imgDB := dbOptions.PickImgDB(curDbId)
		indexToIdentDB := dbOptions.InitIndexDBByBaseDir(curDbId, 3)
		identToIndexDB := dbOptions.InitIndexDBByBaseDir(curDbId, 4)

		dbOptions.ImgIndexSaveRun(curDbId, -1)

		imgDB.CloseDB()
		indexToIdentDB.CloseDB()
		identToIndexDB.CloseDB()
	}
}

