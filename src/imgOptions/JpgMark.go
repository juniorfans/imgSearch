package ImgOptions

import (
	"kmeans"
	"github.com/Comdex/imgo"
	"fmt"
	"strings"
	"imgKmeans"
)

func ImageMarkCentersBlue(fileName string, aroundSize int, threshold, k , factor int)  {
	data,width,height := ImgKmeans.PickPoints(fileName, threshold)

	fmt.Println("kmeas poins: ", len(data))

	wndPoins := ImgKmeans.PickInitCenters(data,width,height,k,factor)

	initCenters := make([]kmeans.Point, len(wndPoins))
	for i, wp := range wndPoins{
		initCenters[i] = wp.Center
		fmt.Println(initCenters[i])
	}


	MarkByPoints(fileName,aroundSize, initCenters)
}

func MarkArround(x, y , around int, r,g,b uint8, data [][][]uint8)  {
	height := len(data)
	width := len(data[0])


	for i := 0;i<=around && x+i < width;i++{

		for j := 0;j<=around && y+j < height;j++{
			markRgb(x+i, y+j, r,g,b, data)
		}

	}
}

func markRgb(x,y int , r,g,b uint8, data [][][]uint8)  {
	//注意 data 数组的形式是 data[height][width][0~3]
	data[y][x][0] = r
	data[y][x][1] = g
	data[y][x][2] = b
}

func MarkByPoints(fileName string, aroundSize int, points []kmeans.Point)  {
	data,err := imgo.Read(fileName)
	if nil != err{
		fmt.Println("open file error: ", fileName, err)
		return
	}

	height := len(data)
	width := len(data[0])


	fmt.Println("height: ", height, ", width: ",width)

	for _, point := range points{
		x := int(point.Entry[0])
		y := int(point.Entry[1])
		//fmt.Println("x: ", x, ", y: ", y)
		MarkArround(x,y,aroundSize,255,0,0, data)
	}

	dst := strings.Replace(fileName, ".", "_mark_.", 1)
	imgo.SaveAsJPEG(dst, data, 100)
}