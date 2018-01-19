package main

import (
	"bufio"
	"os"
	"fmt"
	"dbOptions"
)


func main(){
	stdin := bufio.NewReader(os.Stdin)
	var testCase int
	var dbIndex uint8

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
		/**
		1	有多少小图
		2	某个库下载了多少大图
		3	修复原始图片库/下载库的统计数据
		4	随机校验原始图片库的图片有效性(通过生成图片，肉眼观测)
		5	下载指定库中的图片，通过指定图片的 id 下载
		6	读取切图库中的各个记录的 value(value 是有结构的，key 是子图的 id)
		7	统计切图库的信息(即从各个图片库处理了多少张图片，各到哪个 id 了)
		8	统计大图索引库信息
		9	删除切图库中的统计信息
		10	根据大图索引库，将大图库重复的大图排序在前输出
		11	根据大图索引库，将大图库重复最多的大图保存为 jpg 文件
		12	保存某个大图的汉字
		13	保存某个大图所有的切图为 jpg 文件
		 */
		if(0 == testCase){
			dbOptions.PrintAllStatInfo()
		}else if(1 == testCase){
			dbOptions.HowManyImageClipIndexes()
		}else if(2 == testCase){
			dbOptions.HowManyImages()
		}else if(3 == testCase){
			dbOptions.ImgDBStatRepair(imgDB)
		}else if(4 == testCase){
			dbOptions.RandomVerify()
		}else if(5 == testCase){
			dbOptions.SaveTheInputImg()
		}else if(6 == testCase){
			clipDB:=dbOptions.InitImgClipsReverseIndexDB()
			dbOptions.ReadClipValues()
			clipDB.CloseDB()
		}else if(7 == testCase){
			dbOptions.StatImgClipsInfo()
		}else if(8 == testCase) {
			dbOptions.StatImgIndexesInfo()
		}else if(9 == testCase){
			dbOptions.DeleteStatImgClipsInfo()
		}else if(10 == testCase){
			db:= dbOptions.InitIndexToImgDB()
			dbOptions.SetIndexSortInfo()
			db.CloseDB()
		}else if(11 == testCase){
			db:= dbOptions.InitIndexToImgDB()
			dbOptions.SaveDuplicatedMostImg()
			db.CloseDB()
		}else if(12 == testCase){
			db := dbOptions.InitImgLetterDB()
			dbOptions.SaveLetterOfImg()
			db.CloseDB()
		}else if(13 == testCase){
			dbOptions.TestClipsIndexSaveToJpgFromImgDB()
		}else if(14 == testCase){
			dbOptions.SaveClipsFromClipReverseIndex()
		}else if(15 == testCase){
			dbOptions.TestClipsSaveToJpgFromImgDB()
		}else if(16 == testCase){
			dbOptions.PrintClipIndexBytes()
		}else if(17 == testCase){
			dbOptions.TestFindClipMainImg()
		}else if(18 == testCase){
			dbOptions.TestReadImgDBKey(imgDB)
		}else{
			fmt.Println("invalid options")
		}
		imgDB.CloseDB()
	}

}