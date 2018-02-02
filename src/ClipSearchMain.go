package main

import (
	"bufio"
	"os"
	"fmt"
	"dbOptions"
	"imgIndex"
)

func main()  {
	for{
		stdin := bufio.NewReader(os.Stdin)
		var dbId, which uint8
		var imgId string
		fmt.Print("input dbId,imgId,which to search: ")
		fmt.Fscan(stdin, &dbId, &imgId, &which)

		d1 := dbOptions.InitIndexDBByBaseDir(dbId, 1)
		d2 := dbOptions.InitIndexDBByBaseDir(dbId, 2)

		d3 := dbOptions.InitIndexDBByBaseDir(dbId, 3)
		d4 := dbOptions.InitIndexDBByBaseDir(dbId, 4)


		index := dbOptions.GetImgClipIndexFromClipIdent(dbId,ImgIndex.FormatImgKey([]byte(imgId)),which)
		if 0 == len(index){
			fmt.Println("can't find index for clip")
			continue
		}
		dbOptions.SearchClip(index)


		d1.CloseDB();d2.CloseDB();d3.CloseDB();d4.CloseDB()
	}
}
