package dbOptions

import (
	"config"
	"util"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
)

func FixClipToIndex()  {
	indexToClipDB := InitIndexToClipDB()
	clipToIndexDB := InitClipToIndexDB()

	batch := leveldb.Batch{}

	count := 0
	iter := indexToClipDB.DBPtr.NewIterator(nil, &indexToClipDB.ReadOptions)
	iter.First()

	for iter.Valid(){
		if !fileUtil.BytesStartWith(iter.Key(), config.STAT_KEY_PREX){
			indexBytes := iter.Key()
			clipIdents := iter.Value()
			if 0 != len(clipIdents) % 6{
				fmt.Println("error, clip idents len is not multiple of 6")
				continue
			}
			for i:=0;i < len(clipIdents); i+=6{
				tckey := fileUtil.CopyBytesTo(clipIdents[i : i+6])
				tvalue := fileUtil.CopyBytesTo(indexBytes)

				batch.Put(tckey, tvalue)

				dbId, imgKey, which := ParseAImgClipIdentBytes(tckey)
				show := strconv.Itoa(int(dbId)) + "-" + string(ParseImgKeyToPlainTxt(imgKey)) + "-" + strconv.Itoa(int(which))
				fmt.Println(show)
				count ++
			}
		}
		iter.Next()
	}

	fmt.Println("add ", batch.Len(), " clip ident")

	clipToIndexDB.WriteBatchTo(&batch)
	clipToIndexDB.CloseDB()
	indexToClipDB.CloseDB()
}