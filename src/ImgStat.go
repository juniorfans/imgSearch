package main

import (
	"bufio"
	"os"
	"fmt"
	"dbOptions"
)


func main(){
	stdin := bufio.NewReader(os.Stdin)
	var testCase , dbIndex int

	for{
		fmt.Println("select a imgdb to stat: ")
		fmt.Fscan(stdin, &dbIndex)
		fmt.Println("input option to run: ")
		fmt.Fscan(stdin,&testCase)

		imgDB := dbOptions.PickImgDB(dbIndex)
		if nil == imgDB{
			fmt.Println("open failed, the db is not exsits or open by other process")
			continue
		}

		newBase, lastCores, lastEachTimes, lastCostSecs, lastRemark := dbOptions.GetStatInfo()
		if(1 == testCase){
			fmt.Println("newBase: ", newBase, ", lastCores: ", lastCores, ", lastEachTimes: ", lastEachTimes, ", lastCostSecs: ", lastCostSecs, ", remark: ", lastRemark)
		}else if(2 == testCase){
			dbOptions.HowManyImages()
		}else if(3 == testCase){
			dbOptions.ImgDBStatRepair()
		}else if(4 == testCase){
			dbOptions.RandomVerify()
		}else if(5 == testCase){
			dbOptions.SaveTheInputImg()
		}else{
			fmt.Println("invalid options")
		}

		imgDB.CloseDB()
	}

}