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

//一个子图索引: 一张图片可能有多个子图索引
type SubImgIndex struct {
	IndexUnits      []IndexUnit //一个小图支持多个索引数据

	UnitLength      int         //每个索引单元的字节数

	KeyOfMainImg    []byte      //主图像在 db 中的键值
	DBIdOfMainImg   uint8       //主图像所在的 db
	Which           uint8       //当前小图是主图像的第几幅小图

	cachedFlatBytes []byte
	IsSourceIndex   bool        //当前索引是否是原始索引, 即非分支索引. 一个原始索引对应多个分支索引

	ConfigId        uint8       //使用的切图配置/Letter 配置 id
}

func (this *SubImgIndex) Clone() *SubImgIndex {
	ret := SubImgIndex{}
	ret.IndexUnits = this.IndexUnits		//此处用浅拷贝即可，一旦生成就不再变化
	ret.KeyOfMainImg = this.KeyOfMainImg
	ret.DBIdOfMainImg = this.DBIdOfMainImg
	ret.Which = this.Which
	ret.ConfigId = this.ConfigId
	ret.UnitLength = this.UnitLength

	ret.cachedFlatBytes = this.cachedFlatBytes	//此处用浅拷贝即可，一旦生成就不再变化
/*
	if 0 != len(this.cachedFlatBytes){
		ret.cachedFlatBytes = make([]byte, len(this.cachedFlatBytes))
		copy(ret.cachedFlatBytes, this.cachedFlatBytes)
	}
*/
	ret.IsSourceIndex = this.IsSourceIndex
	return &ret
}

func (this *SubImgIndex) Init(dbId uint8, mainImgKey[]byte, which uint8, unitLength int, clipConfigId uint8)  {
	this.KeyOfMainImg = mainImgKey
	this.DBIdOfMainImg = dbId
	this.Which = which
	this.ConfigId = clipConfigId
	this.UnitLength = unitLength
	this.cachedFlatBytes = nil
	this.IsSourceIndex = false
}

func (this *SubImgIndex) AddIndex(offset int, index IndexData)  {
	this.IndexUnits = append(this.IndexUnits, IndexUnit{Offset:offset, Index: index})
}

func (this *SubImgIndex) Finish() *SubImgIndex {
	return this
}

func (this *SubImgIndex)GetIndexBytesIn3Chanel () []byte {
	return this.getFlatInfo()
}

func (this *SubImgIndex)GetBranchIndexBytesIn3Chanel () (branchIndexes [][]byte){
	sourceIndex := this.getFlatInfo()
	branchIndexes = ClipIndexBranch(sourceIndex)
	return
}

/**
	获得字节数据.
 */
func (this *SubImgIndex)getFlatInfo () []byte {

	if 0 != len(this.cachedFlatBytes){
		return this.cachedFlatBytes
	}

	if len(this.IndexUnits) == 0{
		return nil
	}else{
		totalSize := 0
		for _,curIndex := range this.IndexUnits{
			totalSize += curIndex.Index.GetLength()
		}
		res := make([]byte, totalSize)
		ci := 0
		for _, curIndex := range this.IndexUnits  {
			ci += copy(res[ci:], curIndex.Index.GetBytes())
		}
		if ci != totalSize{
			fmt.Println("getFlatInfo error: ", totalSize, ", ", ci)
			return nil
		}
		this.cachedFlatBytes = ClipIndexSave3Chanel(res)
		return this.cachedFlatBytes
	}
}


func GetFlatIndexBytesFrom(subIndexes []SubImgIndex) []byte {
	if nil == subIndexes ||  0 == len(subIndexes){
		fmt.Println("can't get indexes for image")
		return nil
	}

	clipCount := len(subIndexes)
	//clipCount 张切图，每张切图的索引单元有 len(clipsIndexs[0].IndexUnits) 个，每个索引单元长度是 clipsIndexes[0].UnitLength(即有多少个点)
	// 四个颜色通道，但是索引时只用到3个，所以乘以 3
	estimateSize := clipCount * subIndexes[0].UnitLength * len(subIndexes[0].IndexUnits) * 3
	//	fmt.Println(clipCount , clipsIndexes[0].UnitLength , len(clipsIndexes[0].IndexUnits))
	//	fmt.Println(imgConfig.ClipIndexLength)

	retBytes := make([]byte, estimateSize)
	recvBytes := 0
	for _, clipIndex := range subIndexes {
		clipIndexBytes := clipIndex.GetIndexBytesIn3Chanel()

		copy(retBytes[recvBytes:], clipIndexBytes)

		recvBytes += len(clipIndexBytes)

		if recvBytes > estimateSize{
			fmt.Println("ERROR: estimate of index length cal error")
			return retBytes
		}
	}

	if len(retBytes) % 3 != 0{
		fmt.Println("error, index bytes len is not multiple of 4")
	}

	return retBytes
}

func (this *PointColor) equalsTo(o *PointColor) bool  {
	return this[0]==o[0] && this[1]==o[1] && this[2]==o[2] && this[3]==o[3]
}

func (this *IndexData) GetLength() int {
	var pcs []PointColor
	pcs = *this

	if 0 == len(pcs){
		return 0
	}

	return len(pcs)*4
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
	if count == 0{
		return IndexData{}
	}else if count < 0{
		count = (y1-y0+1)*(x1-x0+1)
	}else{
		//do nothing
	}
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
	data 是整张 image 的结构化数据
	从 (x0,y0) 到 (x1,y1)，一行一行地遍历 data，取出从第 offset 个开始的 count 个数据

	返回的形式是原始三维数据
 */
func GetPartialStructData(data [][][]uint8, x0,y0, x1,y1 int, offset ,count int) [][][]uint8 {
	ret := imgo.New3DSlice( y1-y0+1, x1-x0+1, 4)
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





/**
	data 是一个子图的结构化数据
	将 indexUnit 指示的一个索引单元应用到 data 中

	IndexDataApplyIntoStructedData 与 FromFlatIndexToStructedData 的区别是，当一个子图有多个索引单元时，可以将这些它们
	一次一次叠加到 data 中。而只有一个 index 则可以调用后者快速构建出一个图片结构化数据

 */
func IndexDataApplyIntoSubImg(data [][][]uint8,indexUnit IndexUnit)  {
	count := 0

	height := len(data)
	width := len(data[0])

	var pcs  []PointColor = indexUnit.Index
	realCount := 0
	realCountLimit := len(pcs)
	for j:=0;j < height;j++  {
		for i:=0;i < width;i++  {
			if count >= indexUnit.Offset{
				if realCount>= realCountLimit{
					return
				}
				var colors [4]uint8 = [4]uint8(pcs[realCount])
				data[j][i][0] = colors[0]
				data[j][i][1] = colors[1]
				data[j][i][2] = colors[2]
				data[j][i][3] = colors[3]
				realCount ++
			}
			count ++
		}
	}
	return
}

/**
	从一个索引单元复原出一张图片的结构化数据
 */
func IndexDataConvertToSubImg(height, width int, indexUnit IndexUnit) [][][]uint8 {
	count := 0
	data := imgo.New3DSlice(height, width, 4)
	var pcs  []PointColor = indexUnit.Index
	realCount := 0
	realCountLimit := len(pcs)
	for j:=0;j < height;j++  {
		for i:=0;i < width;i++  {
			if count >= indexUnit.Offset{
				if realCount>= realCountLimit{
					return data
				}
				var colors [4]uint8 = [4]uint8(pcs[realCount])
				data[j][i][0] = colors[0]
				data[j][i][1] = colors[1]
				data[j][i][2] = colors[2]
				data[j][i][3] = colors[3]
				realCount ++
			}
			count ++
		}
	}
	return data
}