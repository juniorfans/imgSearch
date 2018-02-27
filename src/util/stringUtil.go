package fileUtil

import (
	"fmt"
	"sort"
	"bytes"
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

func CopyBytesPrefixTo(src []byte, limit int) []byte {
	if len(src) < limit{
		limit = len(src)
	}
	ret := make([]byte, limit)
	copy(ret, src[:limit])
	return ret
}

func CopyBytesSuffixTo(src []byte, offset int) []byte {
	if len(src) <= offset{
		return nil
	}
	return CopyBytesTo(src[offset:])
}

func BytesLessthan(left, right []byte) bool {
	nsize := len(left)
	if nsize > len(right){
		nsize = len(right)
	}
	for i:=0;i < nsize;i ++{
		if left[i] == right[i]{

		}else{
			return left[i] < right[i]
		}
	}

	return len(left) < len(right)
}

//比较 left 和 right 的前 limit 个字节是否相等.
//异常情况: 两者长度都小于 limit 则直接比较两者是否全等. 否则一个小于 limit 一个大小 limit 则不相等
func BytesEqualPrefix(left, right []byte, limit int) bool {
	if len(left) >= limit && len(right) >= limit{
		return bytes.Equal(left[: limit], right[: limit])
	}else if len(left) < limit && len(right) < limit{
		return bytes.Equal(left, right)
	}else{
		return false
	}
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

//取组合数 C2
func Combine2(data []byte) [] [2]byte {
	if len(data) < 2{
		return nil
	}
	nsize := len(data)
	ret := make([][2]byte, nsize * (nsize-1) / 2)
	ci := 0

	for i,_ := range data{
		for j:=i+1;j<nsize;j ++{
			ret[ci] = [2]byte{data[i], data[j]}
			ci ++
		}
	}
	if ci*2 != nsize*(nsize-1){
		fmt.Println("cal C2 error")
		return nil
	}
	return ret
}

//取排列数 A2
func Arrange2(data []byte) [] [2]byte {
	if len(data) < 2{
		return nil
	}
	nsize := len(data)
	ret := make([][2]byte, nsize * (nsize-1) )
	ci := 0

	for i,_ := range data{
		for j:=i+1;j<nsize;j ++{
			ret[ci] = [2]byte{data[i], data[j]}
			ci ++

			ret[ci] = [2]byte{data[j], data[i]}
			ci ++
		}
	}
	if ci != nsize*(nsize-1){
		fmt.Println("cal A2 error")
		return nil
	}
	return ret
}




type BytesArrayList [][]byte

func (a BytesArrayList) Len() int {
	return len(a)
}
func (a BytesArrayList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

//根据各元素的长度比较排序，逆序
func (a BytesArrayList) Less(i, j int) bool {
	return len(a[i]) > len(a[j])
}

func BytesArraySortByLengthDesc(data [][]byte)  {
	sort.Sort(BytesArrayList(data))
}

var squareMap [256]int
var squareMapInited = false
func InitByteSquareMap()  {
	if squareMapInited{
		return
	}
	for b:=0; b < 256; b ++{
		squareMap[b] = int(b) * int(b)
	}
	squareMapInited = true
}


func ByteSquare(b byte) int {
	return squareMap[int(b)]
}

//计算欧式距离的平方, 即最后一步不开方
func CalEulSquare(left, right []byte) float64 {
	//欧式距离
	sim := 0
	for i:=0;i < len(left);i++{
		if left[i]>right[i]{
			sim += ByteSquare(left[i] - right[i])
		}else if left[i] < right[i]{
			sim += ByteSquare(right[i] - left[i])
		}
	}

	return float64(sim) / float64(len(left))//math.Pow(sim / float64(len(leftIndex)), 0.5)
}