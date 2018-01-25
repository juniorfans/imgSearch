package main

import (
	"dbOptions"
	"fmt"
	"bufio"
	"os"
)

func main()  {
	for{
		stdin := bufio.NewReader(os.Stdin)
		var dbId uint8
		var imgId string
		fmt.Print("input dbId and imgId to cal: ")
		fmt.Fscan(stdin, &dbId, &imgId)

		var whichl, whichr uint8
		fmt.Print("intpu which left and right: ")
		fmt.Fscan(stdin, &whichl, &whichr)

		dbOptions.ExposeCalCollaboratWithEx(dbId, dbOptions.FormatImgKey([]byte(imgId)), whichl, whichr)

		dbOptions.InitIndexToClipDB().CloseDB()
	}

}
