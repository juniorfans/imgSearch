package ImgIndex

import (
	"config"
	"fmt"
)


/**
	branch clip index 由2个字节的统计字节加上branch clip index组成。统计字节目前是方差和平均值
*/
var IMG_INDEX_BYTES_LEN_int = CLIP_INDEX_BYTES_LEN*int(config.CLIP_COUNTS_OF_IMG)
var CLIP_INDEX_STAT_BYTES_LEN int = 2
var CLIP_INDEX_BYTES_LEN int = 72
var CLIP_BRANCH_INDEX_BYTES_LEN int = CLIP_INDEX_BYTES_LEN + CLIP_INDEX_STAT_BYTES_LEN
var CLIP_INDEX_BRANCH_BITS int = 2
var CLIP_INDEX_BRANCH_BOUND uint8 = 10


var CLIP_STAT_INDEX_SOURCE_INDEX_BRANCH_BITS = 0
var CLIP_STAT_INDEX_EACH_OF_MEAN_AND_SD_BRANCH_BITS = 1	//mean 和 sd 各多少位. 位数越多分支越多占据的空间越多查询效率可能会降低
var CLIP_STAT_INDEX_BYTES_LEN int =
	CLIP_STAT_INDEX_SOURCE_INDEX_BRANCH_BITS + 2* CLIP_STAT_INDEX_EACH_OF_MEAN_AND_SD_BRANCH_BITS

var TheclipSearchConf = &config.ClipSearchConf{
	Delta_sd:6.0,
	Delta_mean:2.0,
	Delta_Eul_square:25.0,
	Delta_Eul:5.0,
}
/**
	获得一个 jpg 图像的 index
	内部实现的原理是，获得它里面每张切图的 index, 取长度为: imgConfig.OverrideClipLength, 将其拼接在一起
 */
func GetImgIndexByRGBAData(data [][][]uint8) [] byte {
	height := len(data)
	width := len(data[0])
	imgConfig := config.GetImgConfigBySize(height, width)
	clipConfig := imgConfig.TheClipConfig

	if nil == imgConfig || nil == clipConfig{
		fmt.Println("can't get img config for height: ", height, ", width: ", width)
		return nil
	}

	//自身就是大图，它不属于任何大图，所以 mainImgKey 为 nil
	//注意下面是 imgConfig.OverrideClipLength 此参数必小于 imgConfig.TheClipConfig.ClipLen
	clipsIndexes := GetClipsIndexOfImgEx(data,imgConfig.Id, nil,imgConfig.TheClipConfig.ClipOffsets, imgConfig.OverrideClipLength)

	return GetImgIndexBySubIndexes(clipsIndexes)
}
