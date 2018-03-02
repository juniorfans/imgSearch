package main

import (
	"dbOptions"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func main()  {


	resDB := dbOptions.InitImgAnswerDB()
	clipSameDB := dbOptions.InitCoordinateClipToVTagDB()


	resDB.DBPtr.CompactRange(util.Range{nil,nil})
	clipSameDB.DBPtr.CompactRange(util.Range{nil,nil})

	resDB.CloseDB()
	clipSameDB.CloseDB()


}
