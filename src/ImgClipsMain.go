package main

import (
	"dbOptions"
	"bufio"
	"os"
	"fmt"
	"config"
)

func main()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbIndex uint8
	var input int

	clipConfig := config.GetClipConfigById(0)

	for {
		fmt.Print("select a image db to deal: ")
		fmt.Fscan(stdin, &dbIndex)
		dbOptions.InitImgClipsReverseIndexDB()
		dbOptions.InitImgClipsIndexDB()
		imgDB := dbOptions.PickImgDB(dbIndex)
		if nil == imgDB{
			fmt.Println("open img db failed: ", dbIndex)
			continue
		}
		fmt.Print("input image num for each thread(8 in total) to deal: ")
		fmt.Fscan(stdin, &input)
		dbOptions.BeginImgClipSave(dbIndex,input, clipConfig.ClipOffsets , clipConfig.ClipLengh)

		dbOptions.InitImgClipsReverseIndexDB().CloseDB()
		imgDB.CloseDB()
	}
}