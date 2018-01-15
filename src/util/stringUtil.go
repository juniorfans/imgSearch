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