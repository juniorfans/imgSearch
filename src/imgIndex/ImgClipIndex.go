package ImgIndex

import (
	"github.com/Comdex/imgo"
	"fmt"
	"config"
	"strings"
	"strconv"
)


type ImgClipIndex struct {
	IndexUnits   []IndexUnit //一个小图支持多个索引数据

	UnitLength int           //每个索引单元的字节数

	KeyOfMainImg []byte      //主图像在 db 中的键值
	Which        int         //当前小图是主图像的第几幅小图

	ClipConfigId uint8       //使用的切图配置 id
}

func (this *ImgClipIndex) Init(mainImgKey[]byte, which int, unitLength int, clipConfigId uint8)  {
	this.KeyOfMainImg = mainImgKey
	this.Which = which
	this.ClipConfigId = clipConfigId
	this.UnitLength = unitLength
}

func (this *ImgClipIndex) AddIndex(offset int, index IndexData)  {
	this.IndexUnits = append(this.IndexUnits, IndexUnit{Offset:offset, Index: index})
}

func (this *ImgClipIndex) Finish() *ImgClipIndex {
	return this
}

/**
	获得字节数据. [TODO 此处要优化为获得所有 index 加上元数据信息]
 */
func (this *ImgClipIndex)GetFlatInfo () []byte {
	if len(this.IndexUnits) == 0{
		return nil
	}else{
		return this.IndexUnits[0].Index.GetBytes()
	}
}

/**
	获得 fileName 图片的 clips 切图数据
 */
func GetDataOfClips(fileName string, offset, indexLength int, toSave bool) []ImgData {
	data,err := imgo.Read(fileName)

	if nil != err{
		fmt.Println("open file error: ", fileName, err)
		return nil
	}

	height := len(data)
	width := len(data[0])


	clipConfig := config.GetClipConfigBySize(height, width)
	smallPicLength := clipConfig.SmallPicHeight*clipConfig.SmallPicWidth

	if indexLength > smallPicLength{
		indexLength = smallPicLength
	}

	offsetLimit := smallPicLength-indexLength
	if offset>offsetLimit{
		offset=offsetLimit
	}

	smallPics := make([]ImgData, clipConfig.SmallPicCountInY*clipConfig.SmallPicCountInX)

	xStep := clipConfig.SmallPicWidth+clipConfig.IntervalXBetweenSmallPic
	yStep := clipConfig.SmallPicHeight+clipConfig.IntervalYBetweenSmallPic

	xLimit := width-clipConfig.IntervalXBetweenSmallPic-clipConfig.SmallPicWidth
	yLimit := height-clipConfig.IntervalYBetweenSmallPic-clipConfig.SmallPicHeight
	index := 0
	for i:=clipConfig.StartOffsetX;i<= xLimit; i+=xStep{

		for j:=clipConfig.StartOffsetY;j<=yLimit;j+=yStep{
			//i,j 为左上角

			rightBotomX := i+clipConfig.SmallPicWidth-1
			rightBotomY := j+clipConfig.SmallPicHeight-1
			fmt.Println(i, j,rightBotomX, rightBotomY, " ------ ", xLimit, yLimit, width, height)
			smallPics[index] = getDataFrom(data,i,j,rightBotomX,rightBotomY, offset, indexLength)

			smallFileName := strings.Replace(fileName, ".", "_clip_" + strconv.Itoa(index) + ".", 1)
			if toSave {
				imgo.SaveAsJPEG(smallFileName, smallPics[index], 100)
			}

			index ++
		}

	}
	return smallPics
}

/**
	获得 fileName 图像中的小图的索引数据
	根据此图大小对应的切割配置去切割此图像为多个小图
	取出这些小图各自从 offsetOfClip 开始的 indexLength 个图像数据，返回这些数据
 */
func GetIndexOfClips(fileName string, offsetOfClip, indexLength int) []IndexData {
	data,err := imgo.Read(fileName)

	if nil != err{
		fmt.Println("open file error: ", fileName, err)
		return nil
	}

	height := len(data)
	width := len(data[0])


	clipConfig := config.GetClipConfigBySize(height, width)
	if nil==clipConfig{
		fmt.Println("can't find clipconfig for current img: height: ", height, ", width: ", width)
		return nil
	}
	smallPicLength := clipConfig.SmallPicHeight*clipConfig.SmallPicWidth

	if indexLength > smallPicLength{
		indexLength = smallPicLength
	}

	offsetLimit := smallPicLength-indexLength
	if offsetOfClip >offsetLimit{
		offsetOfClip =offsetLimit
	}

	indexCount := clipConfig.SmallPicCountInY*clipConfig.SmallPicCountInX

	indexes := make([]IndexData, indexCount)

	xStep := clipConfig.SmallPicWidth+clipConfig.IntervalXBetweenSmallPic
	yStep := clipConfig.SmallPicHeight+clipConfig.IntervalYBetweenSmallPic

	xLimit := width-clipConfig.IntervalXBetweenSmallPic-clipConfig.SmallPicWidth
	yLimit := height-clipConfig.IntervalYBetweenSmallPic-clipConfig.SmallPicHeight
	index := 0
	for i:=clipConfig.StartOffsetX;i<= xLimit; i+=xStep{

		for j:=clipConfig.StartOffsetY;j<=yLimit;j+=yStep{
			//i,j 为左上角

			rightBotomX := i+clipConfig.SmallPicWidth-1
			rightBotomY := j+clipConfig.SmallPicHeight-1
			fmt.Println(i, j,rightBotomX, rightBotomY, " ------ ", xLimit, yLimit, width, height)
			indexes[index] = getFlatDataFrom(data,i,j,rightBotomX,rightBotomY, offsetOfClip, indexLength)

			index ++
		}

	}
	return indexes
}

/**
	获得图片 data 的各个切图的索引数据。这些索引数据各自在切图中的偏移是 offsetOfClip，长度是 indexLength

 */
func GetClipsIndexOfImg(data [][][]uint8, mainImgkey []byte, offsetOfClip, indexLength int) []ImgClipIndex {
	height := len(data)
	width := len(data[0])

	clipConfig := config.GetClipConfigBySize(height, width)
	if nil==clipConfig{
		fmt.Println("can't find clipconfig for current img: height: ", height, ", width: ", width)
		return nil
	}
	smallPicLength := clipConfig.SmallPicHeight*clipConfig.SmallPicWidth

	if offsetOfClip <=0{
		offsetOfClip = 0
	}
	if offsetOfClip >= smallPicLength{
		return nil
	}

	if indexLength > smallPicLength-offsetOfClip || indexLength <= 0{
		indexLength = smallPicLength-offsetOfClip
	}

	clipsCount := clipConfig.SmallPicCountInY*clipConfig.SmallPicCountInX

	retIndexes := make([]ImgClipIndex, clipsCount)

	xStep := clipConfig.SmallPicWidth+clipConfig.IntervalXBetweenSmallPic
	yStep := clipConfig.SmallPicHeight+clipConfig.IntervalYBetweenSmallPic

	xLimit := width-clipConfig.IntervalXBetweenSmallPic-clipConfig.SmallPicWidth
	yLimit := height-clipConfig.IntervalYBetweenSmallPic-clipConfig.SmallPicHeight
	index := 0
	clipIndex := 0
	for i:=clipConfig.StartOffsetX;i<= xLimit; i+=xStep{

		for j:=clipConfig.StartOffsetY;j<=yLimit;j+=yStep{
			//i,j 为左上角

			rightBotomX := i+clipConfig.SmallPicWidth-1
			rightBotomY := j+clipConfig.SmallPicHeight-1
			//	fmt.Println(i, j,rightBotomX, rightBotomY, " ------ ", xLimit, yLimit, width, height)

			curIndex := retIndexes[clipIndex]

			curIndex.Init(mainImgkey,clipIndex,indexLength,clipConfig.Id)

			curIndexData := getFlatDataFrom(data,i,j,rightBotomX,rightBotomY, offsetOfClip, indexLength)

			curIndex.AddIndex(offsetOfClip,curIndexData)

			curIndex.Finish()

			retIndexes[clipIndex] = curIndex

			clipIndex++
			index ++
		}

	}

	return retIndexes
}