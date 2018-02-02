package main

import (
	"imgIndex"

	"util"
)

func main()  {
	index := []byte{15, 16, 40, 36, 98, 24, 75, 80}

	ret := ImgIndex.ClipIndexBranch(index)
	for _, br := range ret{
		fileUtil.PrintBytes(br)
	}
}
