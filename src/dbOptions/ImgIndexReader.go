package dbOptions

import (
	"fmt"
	"strings"
)

func ReadImgStat(dbId uint8)  {
	imgIndexDB := InitIndexToImgDB(dbId)
	if nil == imgIndexDB{
		fmt.Println("open img index db failed")
		return
	}
	iterator := imgIndexDB.DBPtr.NewIterator(nil, &imgIndexDB.ReadOptions)
	if iterator.Valid(){
		fmt.Println("invalid iterator ")
	}
	iterator.First()

	singleCount := 0
	multyCount := 0
	for iterator.Valid() {

		imgKeys := iterator.Value()
		//fmt.Println(string(imgKeys))
		imgKeyArray := strings.Split(string(imgKeys), "-")
		if 1 < len(imgKeyArray) {
			multyCount ++
		}else{
			singleCount ++
		}
		iterator.Next()

	}
	iterator.Release()

	fmt.Println("single: ", singleCount, ", multy: ", multyCount)

	return
}