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

	//	newBase, lastCores, lastEachTimes, lastCostSecs, lastRemark := dbOptions.GetStatInfo()
		if(1 == testCase){
			dbOptions.HowManyImageClips()
		}else if(2 == testCase){
			dbOptions.HowManyImages()
		}else if(3 == testCase){
			dbOptions.ImgDBStatRepair()
		}else if(4 == testCase){
			dbOptions.RandomVerify()
		}else if(5 == testCase){
			dbOptions.SaveTheInputImg()
		}else if(6 == testCase){
			clipDB:=dbOptions.InitImgClipsDB()
			dbOptions.ReadClipValues()
			clipDB.CloseDB()
		}else if(7 == testCase){
			dbOptions.StatImgClipsInfo()
		}else if(8 == testCase) {
			dbOptions.StatImgIndexesInfo()
		}else if(9 == testCase){
			dbOptions.DeleteStatImgClipsInfo()
		}else if(10 == testCase){
			db:= dbOptions.InitImgIndexDB()
			dbOptions.SetIndexSortInfo()
			db.CloseDB()
		}else if(11 == testCase){
			db:= dbOptions.InitImgIndexDB()
			dbOptions.SaveDuplicatedMostImg()
			db.CloseDB()
		}else if(12 == testCase){
			db := dbOptions.InitImgLetterDB()
			dbOptions.SaveLetterOfImg()
			db.CloseDB()
		}else if(13 == testCase){
			dbOptions.TestClipsSaveToJpg()
		}else{
			fmt.Println("invalid options")
		}
		imgDB.CloseDB()
	}

}