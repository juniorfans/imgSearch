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
	fmt.Print("select clipIndexToIdent dbs to reference, split by ,: ")
	fmt.Fscan(stdin, &dbIdStrs)
	dbIdStrList := strings.Split(dbIdStrs, ",")
	dbIds := make([]uint8, len(dbIdStrList))
	for i,dbIdStr := range dbIdStrList{
		dbId,_ := strconv.Atoi(dbIdStr)
		dbIds[i] = uint8(dbId)
	}

	dbOptions.MultityInitClipIndexToIdentDBs(dbIds)


	fmt.Print("select img dbs to deal, split by ,: ")
	fmt.Fscan(stdin, &dbIdStrs)
	dbIdStrList = strings.Split(dbIdStrs, ",")
	dbIds = make([]uint8, len(dbIdStrList))
	for i,dbIdStr := range dbIdStrList{
		dbId,_ := strconv.Atoi(dbIdStr)
		dbIds[i] = uint8(dbId)
		dbOptions.InitMuImgToIndexDB(uint8(dbId))
		dbOptions.InitMuIndexToImgDB(uint8(dbId))
	}

	for _, dbId := range dbIds {
		dbOptions.CalCoordinateForDB(dbId, -1)
	}
}
