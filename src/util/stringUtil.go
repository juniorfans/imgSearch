package fileUtil

import (
	"fmt"
	"sort"
)

//left starts with right
func BytesStartWith(left, right []byte) bool {
	if nil == left {
		return nil == right
	}else if nil == right{
		return true
	}else{
		if len(left) < len(right){
			return false
		}
		for i, cr := range right{
			if left[i] != cr{
				return false
			}
		}
		return true
	}
}

func PrintBytes(data []byte)  {
	for _,d := range data{
		fmt.Printf("%d ", d)
	}
	fmt.Println()
}

func PrintBytesLimit(data []byte, limit int)  {
	for _,d := range data{
		if limit == 0{
			break
		}
		limit --
		fmt.Printf("%d ", d)
	}
	fmt.Println()
}

func CopyBytesTo(src []byte) []byte {
	ret := make([]byte, len(src))
	copy(ret, src)
	return ret
}

func MergeBytesTo(target, given *[]byte) {

	if 0 == len(*given){
		return
	}

	ret := make([]byte, len(*target) + len(*given))
	ci := 0
	if 0 != len(*target){
		ci += copy(ret[ci:], *target)
	}

	ci += copy(ret[ci:], *given)

	*target = ret
}

//大端模式: 原值低位在现在字节数组中的高位，注意顺序
func BytesIncrement(srcBytes []byte) bool {

	nsize := len(srcBytes)
	for i:=nsize-1;i>=0;i --{
		if srcBytes[i] < 255{
			srcBytes[i] ++
			return true
		}else{
			srcBytes[i] = 0
		}
	}
	return false	//溢出
}


//left compares to right, 0 is equals, 1 is left > right, other: -1
//if lenth of one less than another, that means it's less.
func BytesCompare(left, right []byte) int8 {
	llen := len(left)
	rlen := len(right)
	if llen!=rlen{
		if llen < rlen{
			return -1
		}else{
			return 1
		}
	}
	for i,_ := range left{
		if left[i] < right[i]{
			return -1
		}else if left[i] > right[i]{
			return 1
		}else{

		}
	}
	return 0
}


type ByteArray []byte

func (a ByteArray) Len() int {
	return len(a)
}
func (a ByteArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
//先比较横坐标，再比较纵坐标
func (a ByteArray) Less(i, j int) bool {
	return a[i] < a[j]
}

func BytesSort(data []byte)  {
	sort.Sort(ByteArray(data))
}

func RemoveDupplicatedBytes(set []byte) []byte {
	mapper := make(map[uint8] byte)
	for _, d := range set{
		mapper[d] = d
	}
	ret := make([]byte, len(mapper))
	ci := 0
	for d,_ := range mapper{
		ret[ci] = d
		ci ++
	}
	mapper = nil
	return ret
}