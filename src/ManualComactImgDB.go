package main

import (
	"fmt"
	"dbOptions"
	"strconv"
	"github.com/syndtr/goleveldb/leveldb/util"
	"bufio"
	"os"
	"strings"
)

func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string
	for  {
		fmt.Print("input db imgs to compact(split by ,): ")
		fmt.Fscan(stdin, &input)
		dbIds := strings.Split(input, ",")
		dbs := make([] *dbOptions.DBConfig, len(dbIds))


		for i,_ := range dbs{

			idbId , err := (strconv.Atoi(dbIds[i]))
			if nil != err{
				fmt.Println("dbid must be int")
				break
			}

			dbs[i]=dbOptions.PickImgDB(uint8(idbId))
			fmt.Println("start to compact db", idbId)
			dbs[i].DBPtr.CompactRange(util.Range{nil,nil})
			fmt.Println("finished compact db", idbId)
			dbs[i].CloseDB()
		}
	}

}