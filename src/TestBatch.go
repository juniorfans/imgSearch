package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"util"
	"fmt"
)

func main()  {
	batch := leveldb.Batch{}
	imgId := []byte{0,0,0,0,0}
	for i:=0;i < 5;i ++{
		fileUtil.BytesIncrement(imgId)
		 batch.Put(imgId, []byte{9,9})
	}

	fmt.Println("batch.size: ", batch.Len())
	fileUtil.PrintBytes(batch.Dump())
}
