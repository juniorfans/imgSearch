package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"strconv"
	"dbOptions"
	"imgIndex"
	"log"
	"runtime/pprof"
	"util"
)



func main()  {
	//TestCoordinate()

	f, err := os.Create("clipCoordinate.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	RealRun()

	pprof.StopCPUProfile()
}

func RealRun()  {
	fileUtil.InitByteSquareMap()

	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string

	fmt.Print("select reference dbs(clipStstIndexToIdent | ImgIdentToIndex |), split by dot: ")
	fmt.Fscan(stdin, &dbIdStrs)
	dbIdStrList := strings.Split(dbIdStrs, ",")

	for _,dbIdStr := range dbIdStrList{
		dbId,_ := strconv.Atoi(dbIdStr)
		dbOptions.InitStatIndexToClipDB(uint8(dbId))
		dbOptions.InitClipToIndexDB(uint8(dbId))
		//dbOptions.InitClipStatIndexToIdentsDB(uint8(dbId)) 	//用于根据输入大图的子图的 stat 数据去查找 clipIdent, 再得到出现的母图
		dbOptions.InitImgToIndexDB(uint8(dbId))//用于：计算得到子图出现在的母图后，需要得到母图的 index 作为键去计算子图出现的相同母图

	}

	fmt.Print("select img dbs to deal, split by ,: ")
	fmt.Fscan(stdin, &dbIdStrs)
	dbIdStrList = strings.Split(dbIdStrs, ",")
	dbIds := make([]uint8, len(dbIdStrList))
	for i,dbIdStr := range dbIdStrList{
		dbId,_ := strconv.Atoi(dbIdStr)
		dbIds[i] = uint8(dbId)
		dbOptions.PickImgDB(uint8(dbId))
	}

	for _, dbId := range dbIds {
		dbOptions.CalCoordinateForDB(dbId, -1)
	}
	dbOptions.FixCoordinateIndexDB()
}

func TestCoordinate()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbIdStrs string
	fmt.Print("select reference dbs(clipIndexToIdent | ImgIdentToIndex | ImgIndexToIdent), split by dot: ")
	fmt.Fscan(stdin, &dbIdStrs)
	dbIdStrList := strings.Split(dbIdStrs, ",")

	for _,dbIdStr := range dbIdStrList{
		dbId,_ := strconv.Atoi(dbIdStr)
		dbOptions.InitStatIndexToClipDB(uint8(dbId))
		dbOptions.InitImgToIndexDB(uint8(dbId))//用于：计算得到子图出现在的母图后，需要得到母图的 index 作为键去计算子图出现的相同母图
	}

	fmt.Print("input img dbId to deal: ")
	var dbId uint8
	fmt.Fscan(stdin, &dbId)
	fmt.Print("input imgNames, split by - : ")
	var imgNameList string
	fmt.Fscan(stdin, &imgNameList)
	imgNames := strings.Split(imgNameList, "-")
	for _,imgName := range imgNames{
		imgKey := ImgIndex.FormatImgKey([]byte(imgName))
		dbOptions.SearchCoordinateForClip(dbId,imgKey)
	}


}