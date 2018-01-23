package ImgIndex

import (
	"fmt"
)

//编辑 clip 的 index, 只保留 RGB 通道，删除 A 通道
//输入要求是 4 通道索引值
func ClipIndexSave3Chanel(clipIndex []byte) [] byte {
	if len(clipIndex) % 4 != 0{
		fmt.Println("clip index length is not multy of 4")
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
/*
	//转换 rgb
	for i, c := range ret{
		if c >= 250{
			ret[i] = 250
		}
		if c%10 < 5{
			ret[i] = (c/10) * 10
		}else{
			ret[i] = (c/10) * 10 + 10
		}
	}
*/
	return ret
}

func makeIndex(src []byte) []byte {
	ret := make([]byte, len(src))
	copy(ret, src)
	return ret
}

//clip index 进行分支, branchBits 表示使用索引的前几位
// 输入的索引要求是 3 通道索引
func ClipIndexBranch(branchBits int, bound uint8, clipIndexIn3Chanel []byte) [][] byte {
	if branchBits <= 0{
		return [][]byte{ClipIndexSave3Chanel(clipIndexIn3Chanel)}
	}
	indexLen := len(clipIndexIn3Chanel)

	if branchBits > indexLen || branchBits > 24 {
		fmt.Println("branch bits too big: ", branchBits)
		return nil
	}

	totalCount := 1
	branchBytes := make([]*ByteBound, branchBits)
	for i:=0;i < branchBits;i ++{
		c := clipIndexIn3Chanel[i]
		branchBytes[i] = GetBound(c, bound)
		if branchBytes[i].isThirdValid(){
			totalCount *= 3
		}else{
			totalCount *= 2
		}
	}

	branchIndex := make([][]byte, totalCount)
	exsitsCount := 0

	for i, branch:=range branchBytes{
		ci := 0	//每轮计算都重新从 0 开始
		//更改已经存在的分支
		{
			if 0 == exsitsCount {
				up := makeIndex(clipIndexIn3Chanel)
				up[i] = branch.up
				branchIndex[ci] = up;ci ++
			}
			for b:=0;b < exsitsCount;b ++{
				up := makeIndex(branchIndex[b])
				up[i] = branch.up
				branchIndex[ci] = up;ci ++
			}

		}

		{
			if 0 == exsitsCount {
				down := makeIndex(clipIndexIn3Chanel)
				down[i] = branch.down
				branchIndex[ci] = down;ci ++
			}

			for b:=0;b < exsitsCount;b ++{
				down := makeIndex(branchIndex[b])
				down[i] = branch.down
				branchIndex[ci] = down;ci ++
			}

		}

		{
			if branch.isThirdValid(){

				if 0 == exsitsCount {
					third := makeIndex(clipIndexIn3Chanel)
					third[i] = branch.third
					branchIndex[ci] = third;ci ++
				}

				for b:=0;b < exsitsCount;b ++{
					third := makeIndex(branchIndex[b])
					third[i] = branch.third

					branchIndex[ci] = third;ci ++
				}
			}

		}
		exsitsCount = ci
	}

	if exsitsCount != totalCount{
		fmt.Println("error: total count is error. totalCount: ", totalCount, ", realCount: ",exsitsCount)
	}
	return branchIndex
}

type ByteBound struct {
	down, up, third uint8
	thirdValid bool
}

func (this *ByteBound) isThirdValid() bool {
	return this.thirdValid
}
func (this *ByteBound) setThirdInValid() {
	this.thirdValid = false
}
func (this *ByteBound) setThirdValid() {
	this.thirdValid = true
}

//获得某个数字以 bound 为基的上下限
//13 的以10为基上限是 20，下限是10，没有 third
//3 以10为基的上限是10，下限是0，没有 third
//30 以10为基的上限是40，下限是20，third 是30
func GetBound(c uint8, bound uint8) *ByteBound {
	ret := ByteBound{}
	ret.setThirdInValid()

	if c < bound{
		ret.down = 0
		ret.up = bound
	}else{
		if 0 == c % bound{
			ret.setThirdValid()
			ret.third = c
			ret.down = c - bound
			ret.up = c + bound
		}else{
			limit := 255/bound * bound
			if limit <= c{
				ret.up = limit
			}else{
				ret.up = c / bound * bound + bound
			}
			ret.down = c / bound * bound
		}
	}
	return &ret
}

func RecoverClipIndex(clipIndex []byte) []byte {
	return nil
}
