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
		0	打印所有库的统计信息
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
		13	在 clip 子图中标记出 clip index
		14	从 indexToClip 库中保存 clip 子图。这样可以将相似 index 的 clip 子图放在一起输出
		15	指定大图，保存它的所有 clip 子图
		16	从 indexToClip 表中取出 clip 的 index 字节,打印为 ycbcr 形式
		17	从 indexToClip 表中直接打印 clip 的 index 字节
		18	从 imgdb 中读取 img key
		19	从 clipToIndex 库中打印出 clip ident
		20	是否能找到 clip 的 index
		21	从 imgToIndex 库中读取键: imgIdent
		22	是否能从 imgToIndex 库中找到 imgident
		23	从 imgDB 中某处开始，导出多少个图
		24	测试 imgDB 库的划分
		25	测试 imgdb 的划分: 总分核对
		 */
		if(0 == testCase){
			dbOptions.PrintAllStatInfo()
		}else if(1 == testCase){
			dbOptions.HowManyImageClipIndexes(dbIndex)
		}else if(2 == testCase){
			dbOptions.HowManyImages()
		}else if(3 == testCase){
			dbOptions.ImgDBStatRepair(imgDB)
		}else if(4 == testCase){
			dbOptions.RandomVerify()
		}else if(5 == testCase){
			dbOptions.SaveTheInputImg()
		}else if(6 == testCase){
			dbOptions.ReadClipValues(dbIndex)
		}else if(7 == testCase){
			dbOptions.StatImgClipsInfo()
		}else if(8 == testCase) {
			dbOptions.StatImgIndexesInfo()
		}else if(9 == testCase){
			dbOptions.DeleteStatImgClipsInfo()
		}else if(10 == testCase){
			dbOptions.SetIndexSortInfo(dbIndex)
		}else if(11 == testCase){
			dbOptions.SaveDuplicatedMostImg(dbIndex)
		}else if(12 == testCase){
			db := dbOptions.InitImgLetterDB()
			dbOptions.SaveLetterOfImg()
			db.CloseDB()
		}else if(13 == testCase){
			dbOptions.MarkClipIndexOnImg(imgDB)
		}else if(14 == testCase){
			dbOptions.SaveClipsFromIndexToClipdb(dbIndex)
		}else if(15 == testCase){
			dbOptions.SaveAllClipsOfImgs()
		}else if(16 == testCase){
			dbOptions.PrintClipIndexInYCBCR()
		}else if(17 == testCase){
			dbOptions.PrintClipIndex()
		}else if(18 == testCase){
			dbOptions.ReadImgKeyFromImgDB(imgDB)
		}else if(19 == testCase){
			dbOptions.PrintClipIdent(dbIndex)
		}else if(20 == testCase){
			dbOptions.CanFindIndexForClip()
		}else if(21 == testCase){
			dbOptions.PrintImgIdent(dbIndex)
		}else if(22 == testCase){
			dbOptions.CanFindImgIdentInImgToIndexDB(dbIndex)
		}else if(23 == testCase){
			dbOptions.SaveImgOffsetAndCount(imgDB)
		}else if(24 == testCase){
			dbOptions.TestSplitTotalCounts(imgDB, 25)
		}else{
			fmt.Println("invalid options")
		}
	//	imgDB.CloseDB()
	}

}