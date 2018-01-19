package dbOptions

import "fmt"

//编辑 clip 的 index
func EditClipIndex(clipIndex []byte) [] byte {
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
	return ret
}

func RecoverClipIndex(clipIndex []byte) []byte {
	return nil
}