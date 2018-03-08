package main

import (
	"imgSearch/src/dbOptions"
	"imgSearch/src/imgIndex"
	"bufio"
	"os"
	"fmt"
)

func main()  {
	test2()
}

func inputTestEx()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId, threshold, offset, limit int
	for{
		fmt.Print("input dbId, threshold, start and offset to classify: ")
		fmt.Fscan(stdin, &dbId, &threshold, &offset, &limit)
		testEx(uint8(dbId), threshold, offset, limit)
	}
}
func testEx(dbId uint8, threshold, offset, limit int)  {
	db := dbOptions.InitClipToIndexDB(dbId)
	iter := db.DBPtr.NewIterator(nil, &db.ReadOptions)
	iter.First()
	for iter.Valid(){
		offset --
		if offset > 0{
			iter.Next()
			continue
		}

		dbOptions.FindAll(threshold, iter.Key())

		limit --
		if limit<=0{
			break
		}
		iter.Next()
	}
}


func test2()  {
	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	clipIdent[0] = 2
	copy(clipIdent[1:], ImgIndex.FormatImgKey([]byte("A0007425")))	//A0000112
	clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = 5	//4
	dbOptions.FindAll(2, clipIdent)
}