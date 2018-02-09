package main

import (
	"dbOptions"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func main()  {


	resDB := dbOptions.InitImgIndexToWhichDB()
	clipSameDB := dbOptions.InitClipSameDB()


	resDB.DBPtr.CompactRange(util.Range{nil,nil})
	clipSameDB.DBPtr.CompactRange(util.Range{nil,nil})

	resDB.CloseDB()
	clipSameDB.CloseDB()


}
