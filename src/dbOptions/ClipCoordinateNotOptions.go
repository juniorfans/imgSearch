package dbOptions

import (
	"imgIndex"
	"fmt"
)

func TestMarkNotClipCoordinate()  {
	imgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
	imgIdent[0] = 2
	copy(imgIdent[1:],ImgIndex.FormatImgKey([]byte("A0000000")))

	MarkNotClipCoordinate(imgIdent, 0, 1)
	fmt.Println(HasMarkedNotClipCoordinate(imgIdent, 0,1))

	fmt.Println(HasMarkedNotClipCoordinate(imgIdent, 0,2))

	MarkNotClipCoordinate(imgIdent, 3, 2)
	fmt.Println(HasMarkedNotClipCoordinate(imgIdent, 2,3))

	MarkNotClipCoordinate(imgIdent, 1, 2)
	fmt.Println(HasMarkedNotClipCoordinate(imgIdent, 4,3))
}

func HasMarkedNotClipCoordinate(imgIdent []byte, left, right uint8) bool {
	notCoordinateDB := InitNotClipCoordinateIndexDB()

	clipIndexes := GetClipIndexBytesOfWhich(imgIdent[0],imgIdent, []uint8{left, right})

	leftStatBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndexes[left])
	rightStatBranchIndexes := ImgIndex.ClipStatIndexBranch(clipIndexes[right])

	queryKey := make([]byte, 2 * ImgIndex.CLIP_STAT_INDEX_BYTES_LEN)
	valueLen := 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)

	index1Start := 0
	index1Limit := index1Start + ImgIndex.CLIP_INDEX_BYTES_LEN
	index2Start := index1Limit + ImgIndex.IMG_CLIP_IDENT_LENGTH
	index2Limit := index2Start + ImgIndex.CLIP_INDEX_BYTES_LEN

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
				if isSameClip(lindex, clipIndexes[left]) && isSameClip(rindex, clipIndexes[right]){
					return true
				}
			}


		}
	}

	return false
}

func MarkNotClipCoordinate(imgIdent []byte, left, right uint8)  {
	notCoordinateDB := InitNotClipCoordinateIndexDB()

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

	cacheLen := valueLen
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

			exsitsValue := notCoordinateDB.ReadFor(queryKey)
			if len(exsitsValue) + valueLen > cacheLen{
				for len(exsitsValue) + valueLen > cacheLen{
					cacheLen *= 2
				}
				newBuff := make([]byte, cacheLen)

				copy(newBuff, valueBuff[:valueLen])
				copy(newBuff[valueLen:], exsitsValue)
				valueBuff = newBuff

			}else{
				if 0 != len(exsitsValue){
					copy(valueBuff[valueLen:], exsitsValue)
				}
			}

			notCoordinateDB.WriteTo(queryKey, valueBuff)
			notCoordinateDB.WriteTo(anotherQueryKey, valueBuff)
		}
	}
}