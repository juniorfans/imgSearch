package ImgIndex

import "imgSearch/src/imgCache"



func removeDuplicateByClipIndexes(clipIndexAndIdents [][]byte) [][]byte {
	if len(clipIndexAndIdents) == 0{
		return nil
	}

	//每个 stat ex index 有重复的认为
	statExIndexesToClip := imgCache.NewMyMap(false)
	for _,clipIndexandIdent := range clipIndexAndIdents{
		index := clipIndexandIdent[:CLIP_INDEX_BYTES_LEN]
		exIndex := flatMeansOfEachSecion(index, CLIP_INDEX_BYTES_LEN/4)	//将 clip index 分为四部分, 计算每部分的均值, 拼接在一起
		statExIndexesToClip.Put(exIndex, clipIndexandIdent)
	}

	var ret [][]byte

	exIndexes := statExIndexesToClip.KeySet()
	for _,exIndex := range exIndexes{
		indexAndIdents := statExIndexesToClip.Get(exIndex)
		if len(indexAndIdents) == 0{
			continue
		}
		for _, indexAndIdent := range indexAndIdents{
			ret = append(ret, indexAndIdent.([]byte))
		}
	}

	return ret
}


//计算 data 中连续 sectinnLen 的区间的均值, 接接在一起
func flatMeansOfEachSecion(data []byte, sectionLen int) []byte {
	if len(data) == 0{
		return nil
	}

	ncount := len(data) / sectionLen
	ret := make([]byte, ncount)
	ci := 0
	for i:=0;i < len(data);i += sectionLen{
		ret[ci] = meanOf(data[i : i+sectionLen])
		ci ++
	}

	return ret
}

func meanOf(data []byte) uint8 {
	if len(data) == 0{
		return 0
	}
	main := 0
	for _,d := range data{
		main += int(d)
	}

	return uint8(main / len(data))
}

