package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"strconv"
	"dbOptions"
	"log"
	"runtime/pprof"
)

func main()  {

	f, err := os.Create("imgIndexMain.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Println("input dbIds, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")
	for _, dbIdStr := range dbIdStrArray{
		dbIdS,_ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)

		imgDB := dbOptions.PickImgDB(curDbId)
		indexToIdentDB := dbOptions.InitMuIndexToImgDB(curDbId)
		identToIndexDB := dbOptions.InitMuImgToIndexDB(curDbId)
		statIndexToIdentDB := dbOptions.InitClipStatIndexToIdentsDB(curDbId)

		dbOptions.ImgIndexSaveRun(curDbId, -1)

		imgDB.CloseDB()
		indexToIdentDB.CloseDB()
		identToIndexDB.CloseDB()
		statIndexToIdentDB.CloseDB()
	}

	pprof.StopCPUProfile()
}

