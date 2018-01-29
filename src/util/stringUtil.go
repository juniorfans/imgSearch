package fileUtil

import "fmt"

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
	ret := make([]byte, len(*target) + len(*given))
	ci := 0
	if 0 != len(*target){
		ci += copy(ret[ci:], *target)
	}
	if 0 != len(*given){
		ci += copy(ret[ci:], *given)
	}
	*target = ret
}
