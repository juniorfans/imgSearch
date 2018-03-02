package main

import (
	"config"
	"bufio"
	"os"
	"fmt"
	"dbOptions"
	"imgIndex"
	"util"
	"github.com/syndtr/goleveldb/leveldb"
)
/**
	根据输入的子图, 插入 index 数据: 同时将 imgKey 和 img-index 微调. 这样做的效果等价于吸收一张子图相同, 但大图不同的图片.
	此工具用于调试
 */
func main()  {
	for{
		stdin := bufio.NewReader(os.Stdin)

		var dbId, which uint8
		var imgId string
		fmt.Print("input dbId,imgId, which to insert: ")
		fmt.Fscan(stdin, &dbId, &imgId, &which)

		imgKey := ImgIndex.FormatImgKey([]byte(imgId))
		imgIdent := make([]byte, ImgIndex.IMG_IDENT_LENGTH)
		imgIdent[0] = byte(dbId)
		copy(imgIdent[1:], imgKey)

		//由指定的子图所在的大图, 复制出一个略微改动 index 的大图, 微调 imgIdent
		imgToIndexDB := dbOptions.InitImgToIndexDB(dbId)
	 	imgIndex := imgToIndexDB.ReadFor(imgIdent)
		imgIndex[0] = 4
		imgIndex[34] = 3
		imgIndex[90] = 90

		newImgIdent := fileUtil.CopyBytesTo(imgIdent)
		newImgIdent[1]=byte('Z')

		imgToIndexDB.WriteTo(newImgIdent, imgIndex)

		indexToImgDB := dbOptions.InitIndexToImgDB(dbId)
		indexToImgDB.WriteTo(imgIndex, newImgIdent)

		//复制子图的索引
		clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
		copy(clipIdent[:5], newImgIdent)
		//clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = byte(which)

		clipConfig := config.GetClipConfigById(0)
		clipSubIndexes := dbOptions.GetDBIndexOfClips(dbOptions.PickImgDB(dbId), imgKey, clipConfig.ClipOffsets, clipConfig.ClipLengh)
		indexToClipIdentBatch := leveldb.Batch{}
		identToClipIndexBatch := leveldb.Batch{}
		for i, clipSubIndex := range clipSubIndexes {
			clipIndex := clipSubIndex.GetIndexBytesIn3Chanel()
			branchIndexes := ImgIndex.ClipIndexBranch(clipIndex)

			clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = uint8(i)

			//往 index-to-ident 中插入 branch
			for _,branch := range branchIndexes{
				indexToClipIdentBatch.Put(branch, clipIdent)
			}
			//往 ident-to-index 中插入
			identToClipIndexBatch.Put(clipIdent , clipIndex)
		}

		dbOptions.InitClipToIndexDB(dbId).WriteBatchTo(&identToClipIndexBatch)
		dbOptions.InitIndexToClipDB(dbId).WriteBatchTo(&indexToClipIdentBatch)



		ndbId,nImgKey := ImgIndex.ParseImgIdenBytes(newImgIdent)
		fmt.Println("has gernerate for ", ndbId, "_", string(ImgIndex.ParseImgKeyToPlainTxt(nImgKey)))
	}
}
