package ImgIndex

import (
	"config"
	"fmt"
)

func GetIndexFor(data [][][]uint8) [] byte {
	height := len(data)
	width := len(data[0])
	imgConfig := config.GetImgConfigBySize(height, width)
	clipConfig := imgConfig.TheClipConfig

	if nil == imgConfig || nil == clipConfig{
		fmt.Println("can't get img config for height: ", height, ", width: ", width)
		return nil
	}

	clipsIndexes := GetClipsIndexOfImg(data,nil,imgConfig.ClipIndexOffset, imgConfig.ClipIndexLength)
	if nil == clipsIndexes ||  0 == len(clipsIndexes){
		fmt.Println("can't get indexes for image")
		return nil
	}

	clipCount := clipConfig.SmallPicCountInX*clipConfig.SmallPicCountInY
	//clipCount 张切图，每张切图的索引单元有 len(clipsIndexs[0].IndexUnits) 个，每个索引单元长度是 clipsIndexes[0].UnitLength，有四个颜色通道，所以乘以 4
	estimateSize := clipCount * clipsIndexes[0].UnitLength * len(clipsIndexes[0].IndexUnits) * 4
//	fmt.Println(clipCount , clipsIndexes[0].UnitLength , len(clipsIndexes[0].IndexUnits))
//	fmt.Println(imgConfig.ClipIndexLength)

	retBytes := make([]byte, estimateSize)
	recvBytes := 0
	for _, clipIndex := range clipsIndexes{
		clipIndexBytes := clipIndex.GetFlatInfo()

		copy(retBytes[recvBytes:], clipIndexBytes)

		recvBytes += len(clipIndexBytes)

		if recvBytes > estimateSize{
			fmt.Println("ERROR: estimate of index length cal error")
			return retBytes
		}
	}
	return retBytes
}