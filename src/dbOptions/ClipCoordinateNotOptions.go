package dbOptions

import (
	"imgIndex"
	"fmt"
	"imgSearch/src/imgCache"
	"bytes"
)

/**
	该库记录两张子图不是主题相似关系. 格式: statindex1 | statIndex2 --> clipIndex1 | clipIdent1 | clipIndex2 | clipIdent2
	在逻辑上与 coordinate 库是相反的

 */


func TestMarkNotSameTopic()  {
	imgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	imgIdent[0] = 2
	copy(imgIdent[1:],ImgIndex.FormatImgKey([]byte("A0000000")))

	fmt.Println("0-1")
	MarkClipsNotSameTopic(imgIdent, 0, 1)
	fmt.Println(HasMarkedClipsNotSameTopic(imgIdent, 0,1))
	MarkClipsNotSameTopicCancel(imgIdent, 1, 0)
	fmt.Println(HasMarkedClipsNotSameTopic(imgIdent, 0,1))

	fmt.Println("0-2")
	fmt.Println(HasMarkedClipsNotSameTopic(imgIdent, 0,2))
	MarkClipsNotSameTopic(imgIdent, 0, 2)
	fmt.Println(HasMarkedClipsNotSameTopic(imgIdent, 0,2))

	fmt.Println("3-2")
	MarkClipsNotSameTopic(imgIdent, 3, 2)
	fmt.Println(HasMarkedClipsNotSameTopic(imgIdent, 2,3))
	MarkClipsNotSameTopicCancel(imgIdent, 2, 3)
	fmt.Println(HasMarkedClipsNotSameTopic(imgIdent, 3,2))

}

func IsNotSameTopicOfClipIndexes(leftClipIndex, rightClipIndex []byte) bool {
	notCoordinateDB := InitNotSameTopicDB()

	leftStatBranchIndexes := ImgIndex.ClipStatIndexBranch(leftClipIndex)
	rightStatBranchIndexes := ImgIndex.ClipStatIndexBranch(rightClipIndex)

	queryKey := make([]byte, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)
	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	index1Start := 0
	index1Limit := index1Start + ImgIndex.CLIP_INDEX_BYTES_LEN
	index2Start := index1Limit + ImgIndex.IMG_CLIP_IDENT_LENGTH
	index2Limit := index2Start + ImgIndex.CLIP_INDEX_BYTES_LEN

	//由于 stat index 是分支的，相同的 clip index 对应了多个 stat index, 下面两个 map 用于减少匹配的次数
	leftNotSame := imgCache.NewMyMap(false)
	rightNotSame := imgCache.NewMyMap(false)

	for _,leftIndex := range leftStatBranchIndexes{

		copy(queryKey, leftIndex)

		for _,rightIndex := range rightStatBranchIndexes{
			copy(queryKey[ImgIndex.CLIP_STAT_INDEX_BYTES_LEN:], rightIndex)

			exsitsValue := notCoordinateDB.ReadFor(queryKey)
			if 0 == len(exsitsValue){
				continue
			}

			if 0 != len(exsitsValue) % valueLen{
				fmt.Println("error, not_clip_coordinate_index db value length is not multple of ", valueLen, ": ", len(exsitsValue))
				continue
			}

			for i:=0;i < len(exsitsValue);i += valueLen{
				group := exsitsValue[i:i+valueLen]
				lindex := group[index1Start:index1Limit]
				rindex := group[index2Start:index2Limit]
				if leftNotSame.Contains(lindex){
					continue
				}
				if rightNotSame.Contains(rindex){
					continue
				}

				lok := isSameClip(lindex, leftClipIndex)
				rok := isSameClip(rindex, rightClipIndex)
				if lok && rok {
					return true
				}else{
					if !lok{
						leftNotSame.Put(lindex, nil)
					}
					if !rok{
						rightNotSame.Put(rindex, nil)
					}
				}
			}


		}
	}

	leftNotSame.Clear()
	rightNotSame.Clear()

	return false
}

func HasMarkedClipsNotSameTopic(imgIdent []byte, left, right uint8) bool {
	clipIndexes := GetClipIndexBytesOfWhich(imgIdent[0],imgIdent, []uint8{left, right})
	return IsNotSameTopicOfClipIndexes(clipIndexes[left], clipIndexes[right])
}

var cachedNotClipCoordinate *imgCache.MyMap = nil
func duplicateMarkCheck(imgIdent []byte, left, right uint8) bool {
	if nil == cachedNotClipCoordinate{
		cachedNotClipCoordinate = imgCache.NewMyMap(false)
	}
	cachedKey := make([]byte, len(imgIdent) + 2)
	ci := copy(cachedKey, imgIdent)
	cachedKey[ci] = left
	ci ++
	cachedKey[ci] = right
	ci ++
	if cachedNotClipCoordinate.Contains(cachedKey){
		return true
	}else{
		cachedNotClipCoordinate.Put(cachedKey, nil)
	}
	return false
}
func removeFromMarkCache(imgIdent []byte, left, right uint8)  {
	cachedKey := make([]byte, len(imgIdent) + 2)
	ci := copy(cachedKey, imgIdent)
	cachedKey[ci] = left
	ci ++
	cachedKey[ci] = right
	ci ++
	cachedNotClipCoordinate.Remove(cachedKey)
}

/**
	标志子图不是主题相似的. 非线程安全
 */
func MarkClipsNotSameTopic(imgIdent []byte, left, right uint8)  {

	if duplicateMarkCheck(imgIdent, left, right){
		fmt.Println("重复标记无效")
		return
	}

	leftClipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(leftClipIdent, imgIdent)
	leftClipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1]=left

	rightClipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(rightClipIdent, imgIdent)
	rightClipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = right


	clipIndexes := GetClipIndexBytesOfWhich(imgIdent[0],imgIdent, []uint8{left, right})

	leftStatBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndexes[left])
	rightStatBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndexes[right])

	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	valueBuff := make([]byte, valueLen)
	ci := 0
	ci += copy(valueBuff[ci:], clipIndexes[left])
	ci += copy(valueBuff[ci:], leftClipIdent)
	ci += copy(valueBuff[ci:], clipIndexes[right])
	ci += copy(valueBuff[ci:], rightClipIdent)

	queryKey := make([]byte, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)
	anotherQueryKey := make([]byte, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)
	for _,leftIndex := range leftStatBranchIndexes{

		copy(queryKey, leftIndex)
		copy(anotherQueryKey[ImgIndex.CLIP_STAT_INDEX_BYTES_LEN:], leftIndex)
		for _,rightIndex := range rightStatBranchIndexes{
			copy(queryKey[ImgIndex.CLIP_STAT_INDEX_BYTES_LEN:], rightIndex)
			copy(anotherQueryKey, rightIndex)

			addValueForKey(queryKey, valueBuff)
			addValueForKey(anotherQueryKey, valueBuff)
		}
	}
}

func addValueForKey(key []byte, toAdd []byte)  {

	notCoordinateDB := InitNotSameTopicDB()
	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	exsitsValue := notCoordinateDB.ReadFor(key)

	fnds := valueExsitsIn(exsitsValue, toAdd)
	if len(fnds) > 1{
		return
	}

	newValue := make([]byte, len(exsitsValue) + valueLen)
	copy(newValue, toAdd)
	if len(exsitsValue) >0 {
		copy(newValue[valueLen:], exsitsValue)
	}

	//fmt.Println("--------------add as belows----------------")
	//fileUtil.PrintBytes(key)
	//fmt.Print(" : ", len(newValue), " : ")
	//fileUtil.PrintBytes(newValue)
	notCoordinateDB.WriteTo(key, newValue)
}


func MarkClipsNotSameTopicCancel(imgIdent []byte, left, right uint8)  {
	removeFromMarkCache(imgIdent, left, right)

	leftClipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(leftClipIdent, imgIdent)
	leftClipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1]=left

	rightClipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(rightClipIdent, imgIdent)
	rightClipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = right


	clipIndexes := GetClipIndexBytesOfWhich(imgIdent[0],imgIdent, []uint8{left, right})

	leftStatBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndexes[left])
	rightStatBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndexes[right])

	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	var toDelete = make([]byte, valueLen)
	{
		ci := 0
		ci += copy(toDelete[ci:], clipIndexes[left])
		ci += copy(toDelete[ci:], leftClipIdent)
		ci += copy(toDelete[ci:], clipIndexes[right])
		ci += copy(toDelete[ci:], rightClipIdent)
	}

	var anotherToDelete = make([]byte, valueLen)
	{
		ci := 0
		ci += copy(anotherToDelete[ci:], clipIndexes[right])
		ci += copy(anotherToDelete[ci:], rightClipIdent)
		ci += copy(anotherToDelete[ci:], clipIndexes[left])
		ci += copy(anotherToDelete[ci:], leftClipIdent)
	}

	queryKey := make([]byte, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)
	anotherQueryKey := make([]byte, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)
	for _,leftIndex := range leftStatBranchIndexes{

		copy(queryKey, leftIndex)
		copy(anotherQueryKey[ImgIndex.CLIP_STAT_INDEX_BYTES_LEN:], leftIndex)
		for _,rightIndex := range rightStatBranchIndexes{
			copy(queryKey[ImgIndex.CLIP_STAT_INDEX_BYTES_LEN:], rightIndex)
			copy(anotherQueryKey, rightIndex)

			deleteValueInValues(queryKey, toDelete)
			deleteValueInValues(anotherQueryKey, anotherToDelete)
		}
	}
}

func valueExsitsIn(exsitsValue []byte, pattern []byte) []int {
	//找出 exsitsValue 中所有的 toDelete, fnds 是找到的下标

	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	var fnds []int
	pos := 0
	for {
		if pos >= len(exsitsValue){
			break
		}
		offset := bytes.Index(exsitsValue[pos:], pattern)
		if offset == -1{
			break
		}

		//offset 是相对于 0 而言, 而不是 pos, pos+offset 表示在 exsitsValue 中的位置
		offset += pos
		fnds = append(fnds, offset)

		pos = (offset + valueLen)
	}
	return fnds
}

func deleteValueInValues(key []byte, toDelete []byte)  {
	notCoordinateDB := InitNotSameTopicDB()
	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	exsitsValue := notCoordinateDB.ReadFor(key)
	if 0 == len(exsitsValue){
		return
	}

	fnds := valueExsitsIn(exsitsValue, toDelete)

	if 0 == len(fnds){
		return
	}

	ci := 0
	finalValue := make([]byte, len(exsitsValue) - len(fnds) * valueLen)
	validPos := 0
	var validPoss []int
	for _,fnd := range fnds{
		if validPos >= len(exsitsValue){
			break
		}
		if validPos == fnd{
			validPos += valueLen
		}else{
			validPoss = append(validPoss, validPos)
			ci += copy(finalValue[ci:], exsitsValue[validPos : validPos + valueLen])
			validPos += valueLen
		}
	}

	//fmt.Println("valid poses: ", validPoss)
	//
	//fmt.Println("--------------cover as belows----------------")
	//fileUtil.PrintBytes(key)
	//fmt.Print(" : ", len(finalValue)," : ")
	//fileUtil.PrintBytes(finalValue)
	//fmt.Println()


	//覆盖
	notCoordinateDB.WriteTo(key, finalValue)
}