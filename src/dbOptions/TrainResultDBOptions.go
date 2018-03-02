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
	"os"
	"bufio"
)




/**
	结果保存：
	[更改]1. 子图的主题 clipStatIndex -> {clipIndex | clipIdent | tagId} 重复结构
	[删除]2.
	[原样]3. 大图 imgIndexBytes --> which array, 计算方法: 一张大图中选择的哪些子图
	[更改]4. 主题相同的子图有哪些 tagId -> {statInex|clipIndex|clipIdent} 重复结构
	[原样]5. tagId --> tagName 及 tagName --> tagId. 训练时会新加入
 */
//----------------------------------------------------------------------------------
//----------------------------------------------------------------------------------


/*
	1. 子图的主题
	格式: clipStatIndex -> {clipIndex | clipIdent | tagId}
	合并旧值的写入方式
*/
var initedClipTagDb map[int] *DBConfig
func InitClipToTagDB() *DBConfig {
	if nil == initedClipTagDb {
		initedClipTagDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedClipTagDb[hash];ok && true == exsitsDB.inited{
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

	retDB.Name = "result/clip_to_tag/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this clip_to_tag db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedClipTagDb[hash] = &retDB

	return &retDB
}

var CLIP_TAG_DB_VALUE_UINT_BYTES_LEN = ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH + TAG_INDEX_LENGTH
var TAG_CLIP_DB_VALUE_UINT_BYTES_LEN = ImgIndex.CLIP_STAT_INDEX_BYTES_LEN + ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH


//格式: clipStatIndex -> {clipIndex | clipIdent | tagId}
func WriteTheClipToTag(imgIdent []byte, clipIndexBytesOfWhich map[uint8] []byte, whiches []uint8, tagId []byte)  {

	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIdent, imgIdent)

	clipTagDB := InitClipToTagDB()

	exsitsClipIndexes := imgCache.NewMyMap(false)

	ctBuffLen := CLIP_TAG_DB_VALUE_UINT_BYTES_LEN
	ctValueBuff := make([]byte, ctBuffLen)

	for i:=0;i < len(whiches);i ++{

		iw := whiches[i]
		clipIndex := clipIndexBytesOfWhich[iw]
		clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = iw
		ci:=0
		ci += copy(ctValueBuff[ci:], clipIndex)
		ci += copy(ctValueBuff[ci:], clipIdent)
		ci += copy(ctValueBuff[ci:], tagId)

		//已处理过
		if exsitsClipIndexes.Contains(clipIndex){
			continue
		}

		has, statIndexes, values := hasClipIndexExsitsInClipTagValue(clipIndex)
		//有相似的子图已写入, 当前不需要写入
		if has{
			continue
		}
		exsitsClipIndexes.Put(clipIndex, nil)

		for i, statIndex := range statIndexes{
			ctRealLen := len(values[i]) + ci
			if ctRealLen > ctBuffLen {
				for ctRealLen > ctBuffLen {
					ctBuffLen *= 2
				}
				newBuff := make([]byte, ctBuffLen)
				copy(newBuff, ctValueBuff[:ci])

				ctValueBuff = newBuff
			}
			if len(values[i]) > 0{
				copy(ctValueBuff[ci:], values[i])
			}
			clipTagDB.WriteTo(statIndex, ctValueBuff[:ctRealLen])
		}
	}
}

func hasClipIndexExsitsInClipTagValue(clipIndex []byte) (exsits bool, statIndexes [][]byte, values [][]byte) {
	clipTagDB := InitClipToTagDB()

	exsits = false
	statIndexes = ImgIndex.ClipStatIndexBranch(clipIndex)
	values = make([] []byte, len(statIndexes))

	notSame := imgCache.NewMyMap(false)

	for c, iStatIndex := range statIndexes{
		exsitsValue := clipTagDB.ReadFor(iStatIndex)
		values[c] = exsitsValue
		if 0 != len(exsitsValue) % CLIP_TAG_DB_VALUE_UINT_BYTES_LEN{
			fmt.Println("error, clip tag db value length is not multple of ", CLIP_TAG_DB_VALUE_UINT_BYTES_LEN, ": ", len(exsitsValue))
			return
		}

		for i:=0;i < len(exsitsValue);i += CLIP_TAG_DB_VALUE_UINT_BYTES_LEN{
			curInfo := exsitsValue[i: i+CLIP_TAG_DB_VALUE_UINT_BYTES_LEN]
			curClipIndex := curInfo[:ImgIndex.CLIP_INDEX_BYTES_LEN]

			if notSame.Contains(curClipIndex){
				continue
			}
			//当前子图的主题已经存在了
			if isSameClip(curClipIndex, clipIndex){
				exsits = true
				return
			}else{
				notSame.Put(curClipIndex, nil)
				continue
			}
		}
	}
	return
}



//--------------------------------------------------------------
/**
	2.主题对应哪些子图
	//格式: tagid -> clipStatIndex | clipIndex | clipIdent
	clipStatIndex 的作用是为了加速查找
	合并旧值的写入方式
 */
var initedTagToClipDb map[int] *DBConfig

func InitTagToClipDB() *DBConfig {
	if nil == initedTagToClipDb {
		initedTagToClipDb = make(map[int] *DBConfig)
	}

	dbId := uint8(0)

	hash := int(dbId)
	if exsitsDB, ok := initedTagToClipDb[hash];ok && true == exsitsDB.inited{
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

	retDB.Name = "result/tag_to_clip/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_to_clip db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagToClipDb[hash] = &retDB

	return &retDB
}
//格式: tagid -> clipStatIndex | clipIndex | clipIdent
func WriteTheTagToClip(imgIdent []byte, clipIndexBytesOfWhich map[uint8] []byte, whiches []uint8, tagId []byte)  {

	clipIdent := make([]byte, ImgIndex.IMG_CLIP_IDENT_LENGTH)
	copy(clipIdent, imgIdent)

	tagToClipDB := InitTagToClipDB()

	exsitsClipIndexes := imgCache.NewMyMap(false)

	ctBuffLen := TAG_CLIP_DB_VALUE_UINT_BYTES_LEN
	ctValueBuff := make([]byte, ctBuffLen)

	for c:=0;c < len(whiches);c ++{
		iw := whiches[c]
		clipIndex := clipIndexBytesOfWhich[iw]
		clipIdent[ImgIndex.IMG_CLIP_IDENT_LENGTH-1] = iw

		ci:=ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
		ci += copy(ctValueBuff[ci:], clipIndex)
		ci += copy(ctValueBuff[ci:], clipIdent)


		//已处理过
		if exsitsClipIndexes.Contains(clipIndex){
			continue
		}

		has, statIndexes, value := hasClipIndexExsitsInTagValue(tagId,clipIndex)
		//有相似的 clip index 被处理过, 当前也不需要处理
		if has{
			continue
		}
		exsitsClipIndexes.Put(clipIndex, nil)

		if len(value) > 0{
			fmt.Println("tag to clip db exsits value length: ", len(value))
		}

		for _, statIndex := range statIndexes{

			copy(ctValueBuff, statIndex)

			ctRealLen := len(value) + ci
			if ctRealLen > ctBuffLen {
				for ctRealLen > ctBuffLen {
					ctBuffLen *= 2
				}
				newBuff := make([]byte, ctBuffLen)
				copy(newBuff, ctValueBuff[:ci])

				ctValueBuff = newBuff
			}

			if len(value) > 0{
				copy(ctValueBuff[ci:], value)
			}

			//此处不能批量写, 因为逻辑上是合并旧值式的写, 若批量写则可能由于批量内部有相同的键值导致最终的值不是累计的
			tagToClipDB.WriteTo(tagId, ctValueBuff[:ctRealLen])
		}
	}
}

func hasClipIndexExsitsInTagValue(tagId []byte, clipIndex []byte) (exsits bool, statIndexes [][]byte, value []byte) {
	tagToClipDB := InitTagToClipDB()

	exsits = false
	statIndexes = ImgIndex.ClipStatIndexBranch(clipIndex)

	value = tagToClipDB.ReadFor(tagId)

	if 0 != len(value) % TAG_CLIP_DB_VALUE_UINT_BYTES_LEN{
		tagName := InitTagIdToNameDB().ReadFor(tagId)
		fmt.Println("error, tag db value length is not multple of ", TAG_CLIP_DB_VALUE_UINT_BYTES_LEN, ": ", len(value), ", ", string(tagName))
		return
	}

	statIndexStart := 0
	statIndexLimit := statIndexStart + ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
	clipIndexStart := statIndexLimit
	clipIndexLimit := clipIndexStart + ImgIndex.CLIP_INDEX_BYTES_LEN

	notSame := imgCache.NewMyMap(false)

	for _, statIndex := range statIndexes{
		for i:=0;i < len(value);i += TAG_CLIP_DB_VALUE_UINT_BYTES_LEN{
			curInfo := value[i: i+TAG_CLIP_DB_VALUE_UINT_BYTES_LEN]
			curStatIndex := curInfo[statIndexStart : statIndexLimit]
			curClipIndex := curInfo[clipIndexStart : clipIndexLimit]

			//此处可以优化为二分搜索. 但考虑到 tagToClip 库数据量不大, 暂时不优化
			if !bytes.Equal(statIndex, curStatIndex){
				continue
			}

			if notSame.Contains(curClipIndex){
				continue
			}

			//当前子图的主题已经存在了
			if isSameClip(curClipIndex, clipIndex){
				exsits = true
				return
			}else{
				notSame.Put(curClipIndex, nil)
			}
		}
	}
	return
}

//---------------------------------------------------------------------------
/*img 被选择的子图
	格式: (img source index bytes) --> which array
*/
var initedImgWhichesDb map[int] *DBConfig
func InitImgAnswerDB() *DBConfig {
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

	retDB.Name = "result/img_answer/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this img_answer db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedImgWhichesDb[hash] = &retDB

	return &retDB
}

func WriteImgWhiches(dbId uint8, imgIdent []byte, whiches []uint8) error {
	//写入 img index ---> whiches
	resDB := InitImgAnswerDB()

	imgIndexDB := InitImgToIndexDB(dbId)

	index := imgIndexDB.ReadFor(imgIdent)
	if 0 == len(index){
		fmt.Println("get img index null: ", getImgNamgeFromImgIdent(imgIdent))
		return errors.New("get img index null: " + getImgNamgeFromImgIdent(imgIdent))
	}
	//[]uint8 造价于 []byte
	resDB.WriteTo(index, whiches)
	return nil
}


//--------------------------------------------------------------
/**
	tag 库
	格式: tag id/index (2 字节长度) --> tag name
 */
var initedTagIndexToNameDb map[int] *DBConfig

func InitTagIdToNameDB() *DBConfig {
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
	fmt.Println("has pick this tag_id_to_name db: ", retDB.Dir)
	retDB.DBPtr,_ = leveldb.OpenFile(retDB.Dir, &retDB.OpenOptions)
	retDB.inited = true

	initedTagIndexToNameDb[hash] = &retDB

	return &retDB
}

var TAG_INDEX_LENGTH = 2
var STAT_MAX_TAG_INDEX_PREFIX = []byte (string(config.STAT_KEY_PREX) + "_MAX_TAG_INDEX")
func WriteATag(tag []byte) error {

	tag = trimLRSpace(tag)

	tagNameToIndexDB := InitTagNameToIdDB()

	exsistsIndex := tagNameToIndexDB.ReadFor(tag)
	//has exsited
	if nil != exsistsIndex{
		return 	errors.New("tag has been exsited")
	}

	tagIndexToNameDB := InitTagIdToNameDB()

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

func InitTagNameToIdDB() *DBConfig {
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

	retDB.Name = "result/tag_name_to_id/data.db"

	retDB.Dir = retDB.initParams.DirBase + "/"  + retDB.Name
	fmt.Println("has pick this tag_name_to_id db: ", retDB.Dir)
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

	db := InitTagNameToIdDB()
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



//----------------------------------------------------------------------------------
func DupmClipsFromTagToClipDB(limit int)  {
	tagToClipDB := InitTagToClipDB()
	tagIdToNameDB := InitTagIdToNameDB()

	statIndexStart := 0
	statIndexLimit := statIndexStart + ImgIndex.CLIP_STAT_INDEX_BYTES_LEN
	clipIndexStart := statIndexLimit
	clipIndexLimit := clipIndexStart + ImgIndex.CLIP_INDEX_BYTES_LEN
	clipIdentStart := clipIndexLimit
	clipIdentLimit := clipIdentStart + ImgIndex.IMG_CLIP_IDENT_LENGTH

	iter := tagToClipDB.DBPtr.NewIterator(nil, &tagToClipDB.ReadOptions)
	iter.First()
	var value []byte
	for iter.Valid(){
		tag := iter.Key()

		if len(tag) != TAG_INDEX_LENGTH{
			continue
		}

		tagName := tagIdToNameDB.ReadFor(tag)

		value = iter.Value()
		if len(value) == 0 || len(value) % TAG_CLIP_DB_VALUE_UINT_BYTES_LEN != 0{
			fmt.Println("error, tag to clip db value length is not multyple of ", TAG_CLIP_DB_VALUE_UINT_BYTES_LEN, " : ", len(value))
			continue
		}

		for i:=0;i < len(value);i += TAG_CLIP_DB_VALUE_UINT_BYTES_LEN{
			curInfo := value[i: i+TAG_CLIP_DB_VALUE_UINT_BYTES_LEN]

			curClipIdent := curInfo[clipIdentStart : clipIdentLimit]
			dbId := curClipIdent[0]
			imgKey := curClipIdent[1:5]
			which := curClipIdent[5]

			indexes := GetDBIndexOfClips(PickImgDB(dbId) , imgKey, []int{-1} ,-1)

			SaveClipsAsJpgWithName("E:/gen/result_verify/", string(tagName), indexes[which])
		}
		limit --
		if limit <=0 {
			break
		}
		iter.Next()
	}
	iter.Release()
}

func TestQueryClipTag()  {
	stdin := bufio.NewReader(os.Stdin)
	for{
		fmt.Print("input a clip ident: ")
		var input string
		fmt.Fscan(stdin, &input)
		clipIdent := parseToClipIdent(input, "_")
		tagId := QueryTagByClipIdent(clipIdent)
		if len(tagId) == 0{
			fmt.Println("can't find tag for: ", input)
		}else{
			tagName := string(InitTagIdToNameDB().ReadFor(tagId))
			fmt.Println("tag for ", input, " is: ", tagName)
		}
	}
}

func QueryTagByClipIdent(clipIdent []byte) []byte {
	clipIndex := InitClipToIndexDB(clipIdent[0]).ReadFor(clipIdent)
	return QueryTagByClipIndex(clipIndex)
}



func QueryTagByClipIndex(clipIndex []byte) []byte {
	clipTagDB := InitClipToTagDB()
	statIndexes := ImgIndex.ClipStatIndexBranch(clipIndex)

	notSame := imgCache.NewMyMap(false)

	for _,statIndex := range statIndexes{
		value := clipTagDB.ReadFor(statIndex)
		for i:=0;i < len(value);i += CLIP_TAG_DB_VALUE_UINT_BYTES_LEN{

			curInfo := value[i: i+CLIP_TAG_DB_VALUE_UINT_BYTES_LEN]

			curClipIndex := curInfo[:ImgIndex.CLIP_INDEX_BYTES_LEN]
			curTag := curInfo[ImgIndex.CLIP_INDEX_BYTES_LEN + ImgIndex.IMG_CLIP_IDENT_LENGTH : ]

			if notSame.Contains(curClipIndex){
				continue
			}
			//当前子图的主题已经存在了
			if isSameClip(curClipIndex, clipIndex){
				return curTag
			}else{
				notSame.Put(curClipIndex, nil)
				continue
			}
		}
	}
	notSame.Clear()

	return nil
}