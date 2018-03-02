package main

import (
	"os"
	"fmt"
	"bufio"
	"strings"
	"strconv"
	"dbOptions"
	"config"
	"log"
	"runtime/pprof"
)

func main()  {

	f, err := os.Create("clipMain.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	clipConfig := config.GetClipConfigById(0)

	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Println("input dbIds, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")
	for _, dbIdStr := range dbIdStrArray{
		dbIdS,_ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)

		clipIdentToIndexDB := dbOptions.InitClipToIndexDB(curDbId)
		clipIndexToIdentMiddleDB := dbOptions.InitIndexToClipMiddleDB(curDbId)
		imgDB := dbOptions.PickImgDB(curDbId)

		fmt.Println("begin to deal imgdb: ", strconv.Itoa(int(curDbId)))
		dbOptions.BeginImgClipSaveEx(curDbId,-1, clipConfig.ClipOffsets , clipConfig.ClipLengh)
		fmt.Println("end of deal imgdb: ", strconv.Itoa(int(curDbId)))

		clipIdentToIndexDB.CloseDB()
		clipIndexToIdentMiddleDB.CloseDB()
		imgDB.CloseDB()
	}

	pprof.StopCPUProfile()

}
