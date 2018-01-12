package ImgIndex

import (
	"config"
	"fmt"
)

/**
	此处是识别的关键
 */
func GetLetterIndexFor(letterConfig *config.LetterConfig,data [][][]uint8, mainImgkey []byte) []SubImgIndex {
	regions := letterConfig.Regions
	if 0 == len(regions) {
		fmt.Println("letter region config is zero, GetLetterIndexFor error.")
		return nil
	}
	letterIndex := make([]SubImgIndex, 1)
	letterIndex[0].Init(mainImgkey,0,letterConfig.IndexLength,letterConfig.Id)
	for _,reg := range regions{
		indexData := getFlatDataFrom(data,reg.LTX, reg.LTY,reg.RBX,reg.RBY, letterConfig.IndexOffset, letterConfig.IndexLength)
		letterIndex[0].AddIndex(letterConfig.IndexOffset,indexData)
	}
	letterIndex[0].Finish()
	return letterIndex
}