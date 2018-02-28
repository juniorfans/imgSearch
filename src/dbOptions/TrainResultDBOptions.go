package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"imgIndex"
	"errors"
	"config"
	"util"
	"github.com/syndtr/goleveldb/leveldb/util"
	"bytes"
	"imgCache"
	"sort"
)


/**
	结果保存：
	[更改]1. 子图的主题 clipStatIndex -> {clipIndex | clipIdent | tagId} 重复结构
	     2.
	[原样]3. 大图 imgIndexBytes --> which array, 计算方法: 一张大图中选择的哪些子图
	[更改]4. 主题相同的子图有哪些 tagId -> {statInex|clipIndex|clipIdent} 重复结构
	5. tagId --> tagName 及 tagName --> tagId. 训练时会新加入
 */
//----------------------------------------------------------------------------------
func DumpSameTagClip(dbId uint8, limit int)  {
	ctDB := InitClipIndexToTagDB()
	tagIdToNameDB := InitTagIndexToNameDB()
	clipIndexToIdDB := InitMuIndexToClipDB(dbId)
	var clipIdent, tagName []byte
	iter := ctDB.DBPtr.NewIterator(nil, &ctDB.ReadOptions)
	iter.First()
	for iter.Valid(){
		branch := iter.Key()

		if len(branch) != ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN{
			continue
		}

		tagId := iter.Value()

		clipIdent = clipIndexToIdDB.ReadFor(branch)
		if 0 == len(clipIdent){
			continue
		}

		for i:=0;i < len(clipIdent);i += 6{
			curIdent := clipIdent[i:i+6]
			tagName = tagIdToNameDB.ReadFor(tagId)
			if 0 == len(tagName){
				continue
			}

			indexes := GetDBIndexOfClips(PickImgDB(dbId) , curIdent[1:len(curIdent)-1], []int{-1} ,-1)

			SaveClipsAsJpgWithName("E:/gen/result_verify/", string(tagName), indexes[curIdent[len(curIdent)-1]])
			if 0 == limit{
				return
			}
			limit --

		}
		iter.Next()
	}
}




//----------------------------------------------------------------------------------



func GetClipIndexBytesOfWhich(dbId uint8, imgIdent []byte, whiches []uint8) map[uint8] []byte {
	clipIdentToIndexDB := InitMuClipToIndexDB(dbId)

	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIdent, imgIdent)

	if 0 == len(whiches){
		whiches = make([]uint8, config.CLIP_COUNTS_OF_IMG)
		for i:=0;i < int(config.CLIP_COUNTS_OF_IMG);i ++{
			whiches[i] = uint8(i)
		}
	}

	clipIndexes := make(map[uint8] []byte)
	for _,which := range whiches{
		clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = byte(which)
		curIndex := clipIdentToIndexDB.ReadFor(clipIdent)
		if 0 == len(curIndex){
			fmt.Println("get clip index null: ", getClipNamgeFromImgIdent(clipIdent))
			return nil
		}
		clipIndexes[which] = curIndex
	}
	return clipIndexes
}


/*
	主题相似的子图
	格式: (branches clipIndexBytes | branches clipIndexBytes) ---> tagIndex
*/
var initedClipSameDb map[int] *DBConfig
func InitClipSameDB() *DBConfig {
	if nil == initedClipSameDb{
		initedClipSameDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipSameDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		if nil == retDB.initParams{
			return nil
		}
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/clip_to_same_clip/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_to_same_clip db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipSameDb[hash] = &retDB

	return &retDB
}

func WriteTheSameClips(dbId uint8, imgIdent []byte, clipIndexBytesOfWhich map[uint8] []byte, whiches []uint8, tagIndex []byte)  {
	clipSameDB := InitClipSameDB()

	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIdent, imgIdent)

	sameBatch := leveldb.Batch{}
	branchLen := ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN
	toAddKey := make([]byte, 2*branchLen)
	toDupKey := make([]byte, 2*branchLen)

	for i:=0;i < len(whiches);i ++{
		iw := whiches[i]
		iIndex := clipIndexBytesOfWhich[iw]
		iBranches := ImgIndex.ClipIndexBranch(iIndex)
		for _, iBranch := range iBranches{
			copy(toAddKey, iBranch)
			for j:=i+1;j < len(whiches);j ++{
				jw := whiches[j]
				jIndex := clipIndexBytesOfWhich[jw]

				jBranches := ImgIndex.ClipIndexBranch(jIndex)
				for _,jBranch := range jBranches{
					copy(toAddKey[branchLen:], jBranch)
					sameBatch.Put(toAddKey, tagIndex)

					//倒置 toAddKey
					copy(toDupKey, toAddKey[branchLen: ])
					copy(toDupKey[branchLen: ], toAddKey[:branchLen])
					sameBatch.Put(toDupKey, tagIndex)
				}
			}
		}
	}

	clipSameDB.WriteBatchTo(&sameBatch)
}

//---------------------------------------------------------------------------
/*img 被选择的子图
	格式: (img source index bytes) --> which array
*/
var initedImgWhichesDb map[int] *DBConfig
func InitImgIndexToWhichDB() *DBConfig {
	if nil == initedImgWhichesDb {
		initedImgWhichesDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedImgWhichesDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/img_index_to_whiches/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this img result db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedImgWhichesDb[hash] = &retDB

	return &retDB
}

func WriteImgWhiches(dbId uint8, imgIdent []byte, whiches []uint8) error {
	//写入 img index ---> whiches
	resDB := InitImgIndexToWhichDB()

	imgIndexDB := InitMuImgToIndexDB(dbId)

	index := imgIndexDB.ReadFor(imgIdent)
	if 0 == len(index){
		fmt.Println("get img index null: ", getImgNamgeFromImgIdent(imgIdent))
		return errors.New("get img index null: " + getImgNamgeFromImgIdent(imgIdent))
	}
	//[]uint8 造价于 []byte
	resDB.WriteTo(index, whiches)
	return nil
}

//----------------------------------------------------------------
/*
	给 clip 打标签
	branches clipIndex --> tag
*/

var initedClipIndexToTagDb map[int] *DBConfig

func InitClipIndexToTagDB() *DBConfig {
	if nil == initedClipIndexToTagDb {
		initedClipIndexToTagDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipIndexToTagDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/clip_index_to_tag/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_index_to_tag db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipIndexToTagDb[hash] = &retDB

	return &retDB
}

//--------------------------------------------------------------
/**
	各个 tag 与哪些 clipIndex 关联
	格式: (tag| branches clipindex) --> nil
 */
var initedTagToClipIndexDb map[int] *DBConfig

func InitTagToClipIndexDB() *DBConfig {
	if nil == initedTagToClipIndexDb {
		initedTagToClipIndexDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagToClipIndexDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/tag_to_clip_index/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_to_clip_index db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagToClipIndexDb[hash] = &retDB

	return &retDB
}

/**
	写入 bracnes clipIndex --> tagId 和 (tagId | branches clipindex) --> nil

 */
func WriteClipTagDB(clipIndexBytesOfWhich map[uint8] []byte, whiches []uint8, tagIndex []byte)  {
	tagToBranchesIndexBatch := leveldb.Batch{}
	branchesIndexToTagIdBatch := leveldb.Batch{}
	branchLen := ImgIndex.CLIP_BRANCH_INDEX_BYTES_LEN
	toAddKey := make([]byte, branchLen + TAG_INDEX_LENGTH)
	copy(toAddKey[0:TAG_INDEX_LENGTH], tagIndex)

	for i:=0;i < len(whiches);i ++{
		iw := whiches[i]
		iIndex := clipIndexBytesOfWhich[iw]
		iBranches := ImgIndex.ClipIndexBranch(iIndex)
		for _, iBranch := range iBranches{
			branchesIndexToTagIdBatch.Put(iBranch, tagIndex)
			copy(toAddKey[TAG_INDEX_LENGTH:], iBranch)

			tagToBranchesIndexBatch.Put(toAddKey, nil)
		}
	}

	InitTagToClipIndexDB().WriteBatchTo(&tagToBranchesIndexBatch)
	InitClipIndexToTagDB().WriteBatchTo(&branchesIndexToTagIdBatch)
}





//--------------------------------------------------------------
/**
	tag 库
	格式: tag id/index (2 字节长度) --> tag name
 */
var initedTagIndexToNameDb map[int] *DBConfig

func InitTagIndexToNameDB() *DBConfig {
	if nil == initedTagIndexToNameDb {
		initedTagIndexToNameDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagIndexToNameDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/tag_id_to_name/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_index_to_name db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagIndexToNameDb[hash] = &retDB

	return &retDB
}

var TAG_INDEX_LENGTH = 2
var STAT_MAX_TAG_INDEX_PREFIX = []byte (string(config.STAT_KEY_PREX) + "_MAX_TAG_INDEX")
func WriteATag(tag []byte) error {

	tag = trimLRSpace(tag)

	tagNameToIndexDB := InitTagNameToIndexDB()

	exsistsIndex := tagNameToIndexDB.ReadFor(tag)
	//has exsited
	if nil != exsistsIndex{
		return 	errors.New("tag has been exsited")
	}

	tagIndexToNameDB := InitTagIndexToNameDB()

	maxTagIndex := tagIndexToNameDB.ReadFor(STAT_MAX_TAG_INDEX_PREFIX)
	if 0 == len(maxTagIndex){
		maxTagIndex = []byte{0,0}
	}
	if !fileUtil.BytesIncrement(maxTagIndex){
		return 	errors.New("tag increment max error")
	}
	tagIndexToNameDB.WriteTo(STAT_MAX_TAG_INDEX_PREFIX, maxTagIndex)
	tagIndexToNameDB.WriteTo(maxTagIndex, tag)
	tagNameToIndexDB.WriteTo(tag, maxTagIndex)
	return nil
}


//--------------------------------------------------------------
/**
	tag 库
	格式: tag name --> tag id
 */
var initedTagNameToIndexDb map[int] *DBConfig

func InitTagNameToIndexDB() *DBConfig {
	if nil == initedTagNameToIndexDb {
		initedTagNameToIndexDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagNameToIndexDb[hash];ok && true == exsitsDB.inited{
		return exsitsDB
	}

	retDB := DBConfig{
		Dir : "",
		DBPtr : nil,
		inited : false,

		Id:dbId,
		Name:"",
		dbType:2,
	}

	if nil == retDB.initParams{
		retDB.initParams = ReadDBConf("conf_result_db.txt")
		retDB.OpenOptions = *getLevelDBOpenOption(retDB.initParams)
		retDB.initParams.PrintLn()
	}

	{
		retDB.ReadOptions = opt.ReadOptions{}
	}
	{
		retDB.WriteOptions = opt.WriteOptions{Sync:false}
	}

	retDB.Name = "result/tag_name_to_index/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_name_to_index db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagNameToIndexDb[hash] = &retDB

	return &retDB
}

type TagInfo struct {
	TagName          []byte
	TagIndex         []byte
	tagPYList        []string
	tagPYSr          []byte
	tagPYFirstLetter []byte
}

func (this *TagInfo) Print()  {
	fmt.Println(string(this.TagName), " | ", string(this.TagIndex), " | ", this.tagPYList, " | ", string(this.tagPYSr), " | ", string(this.tagPYFirstLetter))
}

type TagInfoList []TagInfo

func (a TagInfoList) Len() int {
	return len(a)
}
func (a TagInfoList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
//先比较横坐标，再比较纵坐标
func (a TagInfoList) Less(i, j int) bool {
	return bytes.Compare(a[i].tagPYFirstLetter, a[j].tagPYFirstLetter) < 0
}


func (this *TagInfoList) Print(){
	for _,tag := range *this{
		tag.Print()
	}
}

func (this *TagInfoList) SortByPinYinFirstLetter(){
	sort.Sort(*this)
}

func (this *TagInfoList) IsTagExsits(tagName []byte) bool{
	for _,tag := range *this{
		if bytes.Equal(tag.TagName, tagName){
			return true
		}
	}
	return false
}

func (this *TagInfoList) MustOnlyOneByName(tagName []byte) *TagInfo{
	for _,tag := range *this{
		if bytes.Equal(tag.TagName, tagName){
			return &tag
		}
	}
	return nil

}

func (this *TagInfoList) FindByNameOrPinyin(input []byte) TagInfoList{
	nres := this.findByName(input)
	pres := this.findByPinYin(input)
	totalRet := make([]TagInfo, len(nres) + len(pres))
	ci := 0
	if 0 != len(nres){
		ci += copy(totalRet[ci:], nres)
	}
	if 0!= len(pres){
		ci += copy(totalRet[ci:], pres)
	}
	ret := TagInfoList(totalRet)
	return (&ret).removeDuplicate()

}

func (this *TagInfoList) findByName(name []byte) TagInfoList {
	var ret []TagInfo
	for _,tag := range *this{
		if fileUtil.BytesStartWith(tag.TagName, name){
			ret = append(ret, tag)
		}
	}
	return ret
}

func (this *TagInfoList) removeDuplicate() TagInfoList {
	var ret []TagInfo
	resMap := imgCache.NewMyMap(false)

	for _, tag := range *this{
		resMap.Put(tag.TagIndex, tag)
	}

	keys := resMap.KeySet()
	for _,key := range keys{
		values := resMap.Get(key)
		if len(values) == 1{
			ret = append(ret, values[0].(TagInfo))
		}
	}

	return TagInfoList(ret)
}

func (this *TagInfoList) findByPinYin(py []byte) TagInfoList{
	var ret []TagInfo

	for _,tag := range *this{
		if fileUtil.BytesStartWith(tag.tagPYFirstLetter, py){
			ret = append(ret, tag)
		}else if fileUtil.BytesStartWith(tag.tagPYSr, py){
			ret = append(ret, tag)
		}else {}

	}

	retList := TagInfoList(ret)
	return (&retList).removeDuplicate()
}
func (this *TagInfoList) FindByIndex (index []byte) TagInfoList {
	var ret []TagInfo
	for _,tag := range *this{
		if bytes.Equal(tag.TagIndex, tag.TagIndex){
			ret = append(ret, tag)
		}
	}
	return TagInfoList(ret)
}

func GetAllTagInfos() TagInfoList {
	ret := queryATagInfoByName(nil)
	ret.SortByPinYinFirstLetter()
	return ret
}

func trimLRSpace(input []byte) []byte {
	start := 0
	limit := len(input)
	for i:=0;i<len(input);i++{
		if ' ' == input[i]{
			start = i+1
		}else{
			break
		}
	}

	for j:=len(input)-1;j >= 0;j --{
		if ' ' == input[j]{
			limit = j
		}else{
			break
		}
	}
	if start>=len(input){
		return []byte{}
	}else if limit <=0 {
		return []byte{}
	}
	return input[start : limit]
}

/*
func QueryTagInfosBy(tagNames [][]byte) []TagInfo {
	var ret []TagInfo = nil
	for _,tagName := range tagNames{
		curIndexes := QueryATagInfo(tagName)
		for _,index := range curIndexes{
			ret = append(ret, index)
		}
	}
	return ret
}
*/

func queryATagInfoByName(tagName []byte) TagInfoList {

	tagName = trimLRSpace(tagName)

	var ret []TagInfo = nil

	var start []byte = tagName
	var limit []byte = nil
	if nil != tagName{
		limit = fileUtil.CopyBytesTo(tagName)
		fileUtil.BytesIncrement(limit)
	}

	db := InitTagNameToIndexDB()
	if 0 == len(start){
		start = nil
	}
	if 0 == len(limit){
		limit = nil
	}
	iter := db.DBPtr.NewIterator(&util.Range{Start:start, Limit:limit}, &db.ReadOptions)

	fileUtil.PrintBytes(start)
	fileUtil.PrintBytes(limit)
	//py := pinyingo.NewPy(pinyingo.STYLE_NORMAL, pinyingo.NO_SEGMENT)
	iter.First()
	for iter.Valid(){
		if len(iter.Value()) == TAG_INDEX_LENGTH{
			pyList := []string{""}//py.Convert(string(iter.Key()))
			pyStr := ""
			pyFirstLetter := ""
			for _,p := range pyList{
				pyStr += p
				if len(p) > 0{
					pyFirstLetter += p[0:1]
				}
			}
			ret = append(ret, TagInfo{
				TagName: fileUtil.CopyBytesTo(iter.Key()),
				TagIndex: fileUtil.CopyBytesTo(iter.Value()),
				tagPYList: pyList,
				tagPYSr: []byte(pyStr),
				tagPYFirstLetter:[]byte(pyFirstLetter),
			},
			)
		}
		iter.Next()
	}
	iter.Release()
	return TagInfoList(ret)
}