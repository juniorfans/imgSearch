package main

import (
	"fmt"
	"dbOptions"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func main()  {
	fmt.Println("start to compact reverse clip index")
	dbOptions.InitIndexToClipDB().DBPtr.CompactRange(util.Range{nil,nil})

	fmt.Println("start to compact clip index")
	dbOptions.InitClipToIndexDB().DBPtr.CompactRange(util.Range{nil,nil})

	fmt.Println("all compact finished ~")
}