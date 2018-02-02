package ImgIndex

import (
	"fmt"
	"util"
	"math"
)


var CLIP_INDEX_STAT_BYTES_LEN int = 2
var CLIP_INDEX_BYTES_LEN int = 72
var CLIP_INDEX_BRANCH_BITS int = 2
var CLIP_INDEX_BRANCH_BOUND uint8 = 10

//编辑 clip 的 index, 只保留 RGB 通道，删除 A 通道
//输入要求是 4 通道索引值
func ClipIndexSave3Chanel(clipIndex []byte) [] byte {
	if len(clipIndex) % 4 != 0{
		fmt.Println("clip index length is not multiple of 4")
		return nil
	}

	//获得 rgb
	rgbaGroupCount := len(clipIndex)/4
	ret := make([]byte, 3 * rgbaGroupCount )
	ci := 0
	for i:=0;i < rgbaGroupCount;i ++  {
		rgb := clipIndex[i*4 : (i+1)*4 - 1]
		ci += copy(ret[ci:], rgb)
	}
	return ret
}

//计算索引值的统计数据: 标准差, 平均值, 最大值, 最小值
func ClipIndexStatInfo(indexBytes []byte) (standardDeviation uint8, mean ,minVlue, maxValue uint8) {
	total := 0
	minVlue = 255
	maxValue = 0
	for _,ib := range indexBytes{
		if ib < minVlue{
			minVlue = ib
		}
		if ib > maxValue{
			maxValue = ib
		}
		total += int(ib)
	}
	mean_float64 := float64(total) / float64(len(indexBytes))

	variance := float64(0)
	for _,ib := range indexBytes{
		variance += math.Pow(mean_float64-float64(ib), 2)
	}

	standardDeviation = uint8(math.Pow(variance/float64(len(indexBytes)), 0.5))
	mean = uint8(mean_float64)
	return
}

//clip index 进行分支, branchBits 表示使用索引的前几位
// 输入的索引要求是 3 通道索引
func ClipIndexBranch(clipIndexBytes []byte) [][] byte {
	branchBits := CLIP_INDEX_BRANCH_BITS
	bound := CLIP_INDEX_BRANCH_BOUND
	if branchBits <= 0{
		return [][]byte{}
	}

	clipIndexIn3Chanel:=make([]byte, len(clipIndexBytes) + CLIP_INDEX_STAT_BYTES_LEN)

	sd, mean,_,_ := ClipIndexStatInfo(clipIndexBytes)

	//设置统计数据字节
	clipIndexIn3Chanel[branchBits] = mean
	clipIndexIn3Chanel[branchBits + 1] = sd

	//拷贝固定的部分
	copy(clipIndexIn3Chanel[CLIP_INDEX_STAT_BYTES_LEN+branchBits:], clipIndexBytes)

	indexLen := len(clipIndexBytes)

	if branchBits > indexLen || branchBits > 24 {
		fmt.Println("branch bits too big: ", branchBits)
		return nil
	}

	totalCount := 1
	branchBytes := make([]*ByteBound, branchBits)
	for i:=0;i < branchBits;i ++{
		c := clipIndexBytes[i]
		branchBytes[i] = GetBound(c, bound)
		totalCount *= branchBytes[i].getValidSize()
	}

	branchIndexes := make([][]byte, totalCount)
	exsitsCount := 0

	for i, branch:=range branchBytes{
		ci := 0	//每轮计算都重新从 0 开始
		//更改已经存在的分支

		//do up
		if branch.upValid{
			if 0 == exsitsCount {
				up := fileUtil.CopyBytesTo(clipIndexIn3Chanel)
				up[i] = branch.up
				branchIndexes[ci] = up;ci ++
			}
			for b:=0;b < exsitsCount;b ++{
				up := fileUtil.CopyBytesTo(branchIndexes[b])
				up[i] = branch.up
				branchIndexes[ci] = up;ci ++
			}
		}

		//do down
		if branch.downValid{
			if 0 == exsitsCount {
				down := fileUtil.CopyBytesTo(clipIndexIn3Chanel)
				down[i] = branch.down
				branchIndexes[ci] = down;ci ++
			}

			for b:=0;b < exsitsCount;b ++{
				down := fileUtil.CopyBytesTo(branchIndexes[b])
				down[i] = branch.down
				branchIndexes[ci] = down;ci ++
			}
		}

		//do third
		if branch.thirdValid{
			if 0 == exsitsCount {
				third := fileUtil.CopyBytesTo(clipIndexIn3Chanel)
				third[i] = branch.third
				branchIndexes[ci] = third;ci ++
			}

			for b:=0;b < exsitsCount;b ++{
				third := fileUtil.CopyBytesTo(branchIndexes[b])
				third[i] = branch.third
				branchIndexes[ci] = third;ci ++
			}
		}
		exsitsCount = ci
	}

	if exsitsCount != totalCount{
		fmt.Println("error: total count is error. totalCount: ", totalCount, ", realCount: ",exsitsCount)
	}


	/*	//暂时不进行 format
	for _,branchIndex := range branchIndexes{
		formatBranchIndex(branchBits + CLIP_INDEX_STAT_BYTES_LEN, branchIndex)

	}
	*/
	return branchIndexes
}

func formatBranchIndex(offset int, branchIndex []byte)  {
	for i, c := range branchIndex{
		if i < offset{
			continue
		}
		if c >= 250{
			branchIndex[i] = 250
		}
		if c%10 < 5{
			branchIndex[i] = (c/10) * 10
		}else{
			branchIndex[i] = (c/10) * 10 + 10
		}
	}
}

type ByteBound struct {
	down, up, third uint8
	downValid, upValid, thirdValid bool
}
func (this *ByteBound) setAllInvalid() {
	this.downValid = false
	this.upValid = false
	this.thirdValid = false
}

func (this *ByteBound) getValidSize() int {
	count := 0
	if this.downValid{
		count ++
	}
	if this.upValid{
		count ++
	}
	if this.thirdValid{
		count ++
	}
	return count
}
func (this *ByteBound) setDown(down uint8)  {
	this.down = down
	this.downValid = true
}
func (this *ByteBound) setUp(up uint8)  {
	this.up = up
	this.upValid = true
}
func (this *ByteBound) setThird(third uint8)  {
	this.third = third
	this.thirdValid = true
}

//获得某个数字以 bound 为基的上下限
//13 的以10为基上限是 20，下限是10，没有 third
//3 以10为基的上限是10，下限是0，没有 third
//30 以10为基的上限是40，下限是20，third 是30
//250 以10为基上限是250, 下限是 240
//251 以10为基没有上限, 下限是 250
//总之, 保证输入值最大向下或者向上，最多跨越 10
func GetBound(c uint8, bound uint8) *ByteBound {
	ret := ByteBound{}
	ret.setAllInvalid()

	if c < bound{
		ret.setDown(0)
		ret.setUp(bound)
	}else{
		limit := 255/bound * bound	//对于 bound=10, limit 即为 250
		if c == limit{
			ret.setDown(limit - bound)
			ret.setUp(limit)
		}else if c > limit{
			ret.setDown(limit)
		}else{
			if 0 == c % bound{
				ret.setThird(c)
				ret.setDown(c - bound)
				ret.setUp(c+bound)
			}else{
				ret.setUp(c / bound * bound + bound)
				ret.setDown(c / bound * bound)
			}
		}
	}
	return &ret
}

func RecoverClipIndex(clipIndex []byte) []byte {
	return nil
}
