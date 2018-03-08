package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/util"
	"imgSearch/src/util"
	"bytes"
	"imgCache"
	"imgSearch/src/imgIndex"
	"fmt"
	"imgSearch/src/config"
	"strconv"
)

func FindAll(threshold int, clipIdent []byte)  {
	clipIndex := InitClipToIndexDB(clipIdent[0]).ReadFor(clipIdent)
	res := imgCache.NewMyMap(false)
	dealed := imgCache.NewMyMap(false)

	clipIndexAndIdent := make([]byte, ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIndexAndIdent, clipIndex)
	copy(clipIndexAndIdent[ImgIndex.CLIP_INDEX_BYTES_LEN:], clipIdent)
	FindAllCoordinateForClip(threshold, clipIndexAndIdent, res, dealed, 0)
	DumpAll(threshold, clipIdent, res)
}

func DumpAll(threshold int, clipIdent []byte, res *imgCache.MyMap)  {
	dir := "E:/gen/classify/" + string(ImgIndex.ParseImgKeyToPlainTxt(clipIdent[1:ImgIndex.IMG_IDENT_LENGTH])) /*+ "_" + strconv.Itoa(int(clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1])) */ + "/"
	if 2 > res.KeyCount(){
		return
	}
	resIdents := filterResult(threshold, res)
	if len(resIdents) < 2{
		return
	}

	for _,ident := range resIdents{
		imgKey := ident[1:5]
		fmt.Println(string(ImgIndex.ParseImgKeyToPlainTxt(imgKey)))
		SaveAClipAsJpgFromClipIdent(dir, ident)
	}
}

func filterResult(threshold int, res *imgCache.MyMap) [][]byte {

	caced := imgCache.NewMyMap(false)

	tmp := imgCache.NewMyMap(true)
	keys := res.KeySet()
	for i, one := range keys{
		for j:=i+1;j < len(keys);j ++{
			two := keys[j]

			support := GetCoordinateSupport(one[:ImgIndex.CLIP_INDEX_BYTES_LEN], two[:ImgIndex.CLIP_INDEX_BYTES_LEN])
			if support >= threshold{
				tmp.Put(one, two)
				tmp.Put(two, one)

				onetwo := make([]byte, len(one) + len(two))
				copy(onetwo, one)
				copy(onetwo[len(one):], two)
				caced.Put(onetwo, nil)

				twoone := make([]byte, len(one) + len(two))
				copy(twoone, two)
				copy(onetwo[len(two):], one)
				caced.Put(twoone, nil)
			}
		}
	}

	coord := imgCache.NewMyMap(false)

	onetwo := make([]byte, 2 * (ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH))
	tmp.RangeFuncFor(func(key []byte, values []interface{}) bool{
		//查看 key 这行数据的集合是否两两有协同关系

		for i,value := range values{
			civ := value.([]byte)
			copy(onetwo, civ)
			for j:=i+1;j < len(values);j ++{
				cjv := values[j].([]byte)
				copy(onetwo[len(civ):], cjv)

				//有两个元素没有协同关系
				if !caced.Contains(onetwo){
					return true
				}
			}
		}
		//执行到这儿说明这一行所有元素是两两有协同关系的
		coord.Put(key, nil)

		return true
	})

	lastRes := imgCache.NewMyMap(false)
	coord.RangeFuncFor(func(key []byte, ignoreValues []interface{}) bool{
		values := tmp.Get(key)

		for _,value := range values{
			cv := value.([]byte)
			lastRes.Put(cv[ImgIndex.CLIP_INDEX_BYTES_LEN : ], nil)
		}
		return true
	})

	if lastRes.KeyCount() == 0{
		return nil
	}


	return lastRes.KeySet()
}

func FindAllCoordinateForClip(threshold int, clipIndexAndIdent []byte, res *imgCache.MyMap, dealed *imgCache.MyMap, level int)  {

	if level >= 5{
		return
	}


	if hasInMap(clipIndexAndIdent, dealed){
		return
	}
	dealed.Put(clipIndexAndIdent, nil)

	if !hasInMap(clipIndexAndIdent, res){
		res.Put(clipIndexAndIdent, nil)
	}

	statIndexes :=  ImgIndex.ClipStatIndexBranch(clipIndexAndIdent[:ImgIndex.CLIP_INDEX_BYTES_LEN])

	for _,statIndex := range statIndexes{
		findForStatIndex(threshold,clipIndexAndIdent, statIndex, res, dealed, level)
	}
}


func findForStatIndex(threshold int, targetClipIndexAndIden, statIndex []byte, res *imgCache.MyMap, dealed *imgCache.MyMap, level int)  {

	db := InitCoordinateClipToVTagDB()
	limit := fileUtil.CopyBytesTo(statIndex)
	fileUtil.BytesIncrement(limit)
	r := util.Range{Start:statIndex, Limit:limit}

	valueLen := 2*(ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH) + CLIP_VIRTUAL_TAGID_LEN + 4
	rightStart := ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH
	rightLimit := rightStart + ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH

	iter := db.DBPtr.NewIterator(&r, &db.ReadOptions)
	iter.First()
	for iter.Valid(){
		if !config.IsValidUserDBKey(iter.Key()){
			iter.Next()
			continue
		}

		if !bytes.Equal(iter.Key()[:ImgIndex.CLIP_STAT_INDEX_BYTES_LEN], statIndex){
			break
		}

		value := iter.Value()
		if len(value) % valueLen != 0{
			iter.Next()
			continue
		}

		for i:=0;i < len(value); i+=valueLen{
			group := value[i: i+valueLen]

			supportBytes := group[valueLen-4:]
			support := ImgIndex.BytesToInt32(supportBytes)
			if support < threshold{
				continue
			}

			left := group[:rightStart]
			//如果当前单元的左子图与 target 是相似子图则继续处理
			if !isSameClip(left[:ImgIndex.CLIP_INDEX_BYTES_LEN], targetClipIndexAndIden[:ImgIndex.CLIP_INDEX_BYTES_LEN]){
				continue
			}

			//左子图相似，则右子图要作为协同子图加进来. 此处可能要优化为迭代计算加入右子图的
			right := group[rightStart : rightLimit]

			lname := ImgIndex.FromClipIdentsToStrings(left[ImgIndex.CLIP_INDEX_BYTES_LEN: ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH])[0]
			rname := ImgIndex.FromClipIdentsToStrings(right[ImgIndex.CLIP_INDEX_BYTES_LEN: ImgIndex.CLIP_INDEX_BYTES_LEN+ImgIndex.IMG_CLIP_IDENT_LENGTH])[0]
			p := string(lname + " - " + rname + " : " + strconv.Itoa(support))
			fmt.Println(p)

			if !hasInMap(right, res){
				res.Put(right, nil)

				FindAllCoordinateForClip(threshold, right,res,dealed, level+1)

			}
		}

		iter.Next()
	}
}


/*
func hasInMap(indexAndIdent []byte, res *imgCache.MyMap) bool {
	fnd := false
	res.RangeFuncFor(func(key []byte, values []interface{} ) bool {
		if isSameClip(key[:ImgIndex.CLIP_INDEX_BYTES_LEN], indexAndIdent[:ImgIndex.CLIP_INDEX_BYTES_LEN]){
			fnd = true
			return false	//不需要继续访问了
		}else{
			return true
		}
	})
	return fnd
}

*/