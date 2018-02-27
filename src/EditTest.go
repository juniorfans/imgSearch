package main

import (
	"imgIndex"

	"util"
	"time"
	"fmt"
)

func main()  {
	fileUtil.InitByteSquareMap()
	newTest()
}

func oldTest()  {
	index := []byte{15, 16, 40, 36, 98, 24, 75, 80}

	ret := ImgIndex.ClipIndexBranch(index)
	for _, br := range ret{
		fileUtil.PrintBytes(br)
	}
}

func newTest()  {
	index := []byte{15, 16, 40, 36, 98, 24, 75, 80}

	startT := time.Now().UnixNano()
	ret := ImgIndex.ClipStatIndexBranch(index)
	for _, br := range ret{
		fileUtil.PrintBytes(br)
	}
	endT := time.Now().UnixNano()

	fmt.Println("cost : ", (endT-startT)/1000000, " ms")
}