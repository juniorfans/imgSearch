package main

import (
	"dbOptions"
	"imgIndex"
	"bufio"
	"os"
	"fmt"
	"util"
)

func main()  {
	TestReadImg()
}

func TestReadClipIndex()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId, which uint8
	var imgKey string

	dbOptions.InitClipToIndexDB(2)
	dbOptions.InitClipToIndexDB(22)
	dbOptions.InitClipToIndexDB(26)

	dbs := dbOptions.GetInitedClipIdentToIndexDB()

	fmt.Println("picked: ")
	for _,db := range dbs{
		fmt.Print(db.Name, "_" ,db.Id, " | ")
	}
	fmt.Println()
	multyDBReader := dbOptions.NewMultyDBReader(dbs, false)

	for{
		fmt.Print("input dbId, imgKey, which: ")
		fmt.Fscan(stdin,&dbId, &imgKey, &which)
		clipIdent := ImgIndex.GetImgClipIdent(dbId, ImgIndex.FormatImgKey([]byte(imgKey)), which)
		indexes,_ := multyDBReader.ReadFor(clipIdent)
		fmt.Println("result: ")
		for _,index := range indexes{
			fileUtil.PrintBytes(index)
		}
	}

	multyDBReader.Close()
}

func TestReadImg()  {
	stdin := bufio.NewReader(os.Stdin)
	var imgKey string

	dbOptions.PickImgDB(2)
	dbOptions.PickImgDB(26)
	dbOptions.PickImgDB(27)

	dbs := dbOptions.GetImgDBs()

	fmt.Println("picked: ")
	for _,db := range dbs{
		fmt.Print(db.Name, "_" ,db.Id, " | ")
	}
	fmt.Println()
	multyDBReader := dbOptions.NewMultyDBReader(dbs, false)

	for{
		fmt.Print("input imgKey: ")
		fmt.Fscan(stdin,&imgKey)

		imgBytes,_ := multyDBReader.ReadFor(ImgIndex.FormatImgKey([]byte(imgKey)))
		fmt.Println("result: ")
		for _,img := range imgBytes{
			fmt.Println(imgKey, " length: ", len(img))
		}
	}

	multyDBReader.Close()
}