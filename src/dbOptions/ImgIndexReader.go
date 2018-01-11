package dbOptions

import (
	"fmt"
	"strings"
)

func ReadImgIndex(count int)  {
	imgIndexDB := InitImgIndexDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db failed")
		return
	}
	iterator := imgIndexDB.DBPtr.NewIterator(nil, &imgIndexDB.ReadOptions)
	if iterator.Valid(){
		fmt.Println("invalid iterator ")
	}
	iterator.First()

	for iterator.Valid() {
		if count == 0 {
			break
		}
		imgKeys := iterator.Value()
		//fmt.Println(string(imgKeys))
		imgKeyArray := strings.Split(string(imgKeys), "-")
		if 1 < len(imgKeyArray) {
			fmt.Println(imgKeyArray)
		}
		iterator.Next()
		count --
	}
	iterator.Release()
	return
}

func ReadImgStat()  {
	imgIndexDB := InitImgIndexDB()
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