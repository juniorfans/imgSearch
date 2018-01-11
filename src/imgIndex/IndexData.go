package ImgIndex

import (
	"github.com/Comdex/imgo"
	"fmt"
)

//一张 image的源数据
type ImgData [][][]uint8

//图中一个像素点的颜色信息
type PointColor  [4]uint8

//image 中切图/ 或者 image 的索引数据
type IndexData  []PointColor

//一个索引单元：从 Offset 开始的索引数据
type IndexUnit struct {
	Offset int                 	//索引数据在小图中的偏移
	Index IndexData 		//当前小图的索引数据
}

func (this *PointColor) equalsTo(o *PointColor) bool  {
	return this[0]==o[0] && this[1]==o[1] && this[2]==o[2] && this[3]==o[3]
}


func (this *IndexData)GetBytes() []byte {

	var pcs []PointColor
	pcs = *this

	if 0 == len(pcs){
		return nil
	}

	ret := make([]byte, len(pcs)*4)
	i := 0
	for _,pc := range pcs{
		for _,c := range pc{
			ret[i]=c
			i++
		}
	}
	return ret
}

/**
	从 (x0,y0) 到 (x1,y1)，一行一行地遍历 data，取出从第 offset 个开始的 count 个数据

	返回的形式是平坦的 PointColor 数组，也即 IndexData
 */
func getFlatDataFrom(data [][][]uint8, x0,y0,x1,y1 int, offset, count int) IndexData {
	ret := make([]PointColor, count)
	realIc := 0
	ic := 0
	for j:=y0;j<=y1;j++{

		for i:=x0;i<=x1 ;i++ {
			if ic >= offset{
				ret[realIc] = PointColor{data[j][i][0],data[j][i][1],data[j][i][2],data[j][i][3]}

				realIc ++
			}
			ic ++
			if realIc == count{
				return IndexData(ret)
			}
		}
	}
	fmt.Println("Warning, no enough data to get, diff: ", count-realIc)
	return IndexData(ret)
}


/**
	从 (x0,y0) 到 (x1,y1)，一行一行地遍历 data，取出从第 offset 个开始的 count 个数据

	返回的形式是原始三维数据
 */
func getDataFrom(data [][][]uint8, x0,y0, x1,y1 int, offset ,count int) [][][]uint8 {
	ret := imgo.New3DSlice(x1-x0+1, y1-y0+1,4)
	realIc := 0
	ic := 0
	for j:=y0;j<=y1;j++{
		iy := j-y0
		for i:=x0;i<=x1 ;i++ {
			ix := i-x0
			if ic >= offset{
				ret[iy][ix][0]=data[j][i][0]
				ret[iy][ix][1]=data[j][i][1]
				ret[iy][ix][2]=data[j][i][2]
				ret[iy][ix][3]=data[j][i][3]

				realIc ++
			}

			ic ++
			if realIc == count{

				return ret
			}
		}
	}
	return ret
}


