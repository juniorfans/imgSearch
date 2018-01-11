package ImgOptions

import (
	"github.com/Comdex/imgo"
	"fmt"
	"sort"
	"strings"
)

func rGB2Gray(src [][][]uint8) [][][]uint8 {
	height := len(src)
	width := len(src[0])
	imgMatrix := imgo.NewRGBAMatrix(height,width)
	copy(imgMatrix,src)

	for i:=0;i<height;i++{
		for j:=0;j<width;j++{
			avg:=uint8(int(imgMatrix[i][j][0])+int(imgMatrix[i][j][1])+int(imgMatrix[i][j][2])/3)
			imgMatrix[i][j][0] = avg
			imgMatrix[i][j][1] = avg
			imgMatrix[i][j][2] = avg
		}
	}
	return imgMatrix
}

type PointStat struct{
	x,y int	//(x,y) 表示一个 point
	grayValue uint8
}
type PointStatArray []PointStat

func (a PointStatArray) Len() int {
	return len(a)
}
func (a PointStatArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}


//先比较横坐标，再比较纵坐标
func (a PointStatArray) Less(i, j int) bool {

	return a[i].grayValue < a[j].grayValue
}



func Statistics(fileName string, threshold int)  {
	data,err := imgo.Read(fileName)
	if nil != err{
		fmt.Println("open file error: ", fileName, err)
		return
	}

	data = rGB2Gray(data)

	height := len(data)
	width := len(data[0])

	pointStats := make([]PointStat, height * width)

	index := 0
	for j:=0;j != height;j ++{
		for i:=0;i != width;i ++{
			pointStats[index].x = i
			pointStats[index].y = j
			pointStats[index].grayValue = data[j][i][0]
			index ++
		}
	}
	sort.Sort(PointStatArray(pointStats))

	cur := pointStats[0].grayValue
	count := 0
	for _,ps :=range pointStats{
		if ps.grayValue==cur{
			count++
		}else{
			fmt.Println(cur," -- ", count)
			cur = ps.grayValue
			count = 1
		}
	}


	for j:=0;j != height;j ++{
		for i:=0;i != width;i ++{
			if data[j][i][0] > uint8(threshold){
				data[j][i][0] = 255
				data[j][i][1] = 255
				data[j][i][2] = 255
			}
		}
	}
	dst := strings.Replace(fileName, ".", "_statistcs_.", 1)
	imgo.SaveAsJPEG(dst, data, 100)
}
