package main

import (
	"dbOptions"
	"fmt"
	"bufio"
	"os"
	"imgIndex"
)

func main()  {
	for{
		stdin := bufio.NewReader(os.Stdin)
		var dbId uint8
		var imgId string
		fmt.Print("input dbId and imgId to cal: ")
		fmt.Fscan(stdin, &dbId, &imgId)

		d1 := dbOptions.InitIndexDBByBaseDir(dbId, 1)
		d2 := dbOptions.InitIndexDBByBaseDir(dbId, 2)

		d3 := dbOptions.InitIndexDBByBaseDir(dbId, 3)
		d4 := dbOptions.InitIndexDBByBaseDir(dbId, 4)

		var whichl, whichr uint8
		fmt.Print("intpu which left and right: ")
		fmt.Fscan(stdin, &whichl, &whichr)

		dbOptions.ExposeCalCollaboratWithEx(dbId, ImgIndex.FormatImgKey([]byte(imgId)), whichl, whichr)


		d1.CloseDB();d2.CloseDB();d3.CloseDB();d4.CloseDB()
	}

}
