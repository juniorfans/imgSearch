package main

import (
	"bufio"
	"os"
	"fmt"
	"dbOptions"
)

func SaveMainImgIndexes()  {
	stdin := bufio.NewReader(os.Stdin)
	var input int
	var dbIndex uint8

	for{
		fmt.Println("select a image db to deal: ")
		fmt.Fscan(stdin,&dbIndex)
		imgDB := dbOptions.PickImgDB(dbIndex)
		indexToImgDB := dbOptions.InitMuIndexToImgDB(dbIndex)
		imgToIndexDB := dbOptions.InitMuImgToIndexDb(dbIndex)
		fmt.Println("input how many times each thread(16 in total) to deal: ")
		fmt.Fscan(stdin,&input)
		dbOptions.ImgIndexSaveRun(dbIndex, input)
		imgDB.CloseDB()
		indexToImgDB.CloseDB()
		imgToIndexDB.CloseDB()
	}
}

func main(){
	SaveMainImgIndexes()
}