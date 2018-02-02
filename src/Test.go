package main

import (
	"fmt"
	"imgIndex"
	"util"
)

func main()  {
	testBytes := []byte{170,201,146,139,168,120,126,168,118,126,168,118,241,206,72,246,211,77,136,128,141,112,106,114,166,203,146,175,211,151,22,22,34,27,27,39,171,206,160,165,199,150,165,202,143,166,203,144}

	standardDeviation,mean, min, max := ImgIndex.ClipIndexStatInfo(testBytes)

	fmt.Println("sd: ", standardDeviation, ", mean: " ,mean, ", min: ", min, ", max: ", max)

	branchIndexes := ImgIndex.ClipIndexBranch(testBytes)
	for _, bindex := range branchIndexes{
		fileUtil.PrintBytes(bindex)
	}


	srcBytes := []byte{1,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255}
	fileUtil.BytesIncrement(srcBytes)
	fileUtil.PrintBytes(srcBytes)
}
