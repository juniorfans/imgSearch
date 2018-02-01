package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"strconv"
	"dbOptions"
)

func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Println("input dbIds, split by ,")
	fmt.Fscan(stdin,&dbIdStrs)
	dbIdStrArray := strings.Split(dbIdStrs, ",")

	for _, dbIdStr := range dbIdStrArray {
		dbIdS, _ := strconv.Atoi(dbIdStr)
		curDbId := uint8(dbIdS)
		dbOptions.MergeTo(dbOptions.PickImgDB(curDbId), dbOptions.PickImgDB(250))
	}
}
