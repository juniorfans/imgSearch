package main

import (
	"bufio"
	"os"
	"fmt"
	"dbOptions"
	"imgIndex"
	"strings"
	"strconv"
	"config"
)

func main()  {
	TestSeek3()
}

func TestSeek3()  {
	for{
		config.MustReReadSearchConf("clip_search_conf.txt")

		stdin := bufio.NewReader(os.Stdin)

		var indexDbIdStr string
		fmt.Print("input index db, split by dot(,): ")
		fmt.Fscan(stdin, &indexDbIdStr)
		indexDBIds := strings.Split(indexDbIdStr, ",")
		for _,indexDBId := range indexDBIds{
			dbId, _ := strconv.Atoi(indexDBId)
			dbOptions.InitMuIndexToClipDB(uint8(dbId))
		}

		var dbId uint8
		var imgId string
		fmt.Print("input dbId,imgId to search: ")
		fmt.Fscan(stdin, &dbId, &imgId)
		dbOptions.SearchClipsOfImg(dbId, ImgIndex.FormatImgKey([]byte(imgId)))
	}
}

func TestSeek1()  {
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

func TestSeek2()  {
	for{
		config.MustReReadSearchConf("clip_search_conf.txt")

		stdin := bufio.NewReader(os.Stdin)

		var indexDbIdStr string
		fmt.Print("input index db, split by dot(,): ")
		fmt.Fscan(stdin, &indexDbIdStr)
		indexDBIds := strings.Split(indexDbIdStr, ",")
		for _,indexDBId := range indexDBIds{
			dbId, _ := strconv.Atoi(indexDBId)
			dbOptions.InitMuIndexToClipDB(uint8(dbId))
			dbOptions.InitMuImgToIndexDB(uint8(dbId))
		}

		var dbId, which uint8
		var imgId string
		fmt.Print("input dbId,imgId,which to search: ")
		fmt.Fscan(stdin, &dbId, &imgId, &which)

		index := dbOptions.GetImgClipIndexFromClipIdent(dbId,ImgIndex.FormatImgKey([]byte(imgId)),which)
		if 0 == len(index){
			fmt.Println("can't find index for clip")
			continue
		}
		dbOptions.SearchClipEx(index)
	}
}