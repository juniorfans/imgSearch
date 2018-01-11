package main

import (
	"bufio"
	"os"
	"fmt"
	"dbOptions"
)

func SaveMainImgIndexes()  {
	stdin := bufio.NewReader(os.Stdin)
	var input ,dbIndex int

	for{
		fmt.Println("select a image db to deal: ")
		fmt.Fscan(stdin,&dbIndex)
		dbOptions.PickImgDB(dbIndex)

		fmt.Println("input how many times each thread(8 in total) to deal: ")
		fmt.Fscan(stdin,&input)
		dbOptions.ImgIndexSaveRun(dbIndex, input)
	}
}

func main(){
	SaveMainImgIndexes()
}