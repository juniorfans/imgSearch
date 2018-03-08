package dbOptions

import (
)
import (
	"imgSearch/src/imgIndex"
	"imgSearch/src/config"
	"fmt"
	"imgSearch/src/imgCache"
	"sort"
	"bufio"
	"os"
	"strings"
	"strconv"
	"imgSearch/src/util"
)

type RecognitionResult struct {
	directAnsFnd 	bool
	sameTagFnd	bool
	coordinateFnd	bool

	directAns	[]uint8
	sameTagAns	[][]uint8
	coordinateAns	[][]uint8

	notSameTopic	[][]uint8
}

func (this *RecognitionResult)ToString() string {
	ret := ""
	if this.directAnsFnd{
		ret += fmt.Sprintf("direct answer: %s", fileUtil.UintsToString(this.directAns))
	}else if this.sameTagFnd{
		ret += fmt.Sprintf("same tag: %s", fileUtil.UintsListToString(this.sameTagAns))
	}else if this.coordinateFnd{
		ret += fmt.Sprintf("coordinate: %s", fileUtil.UintsListToString(this.coordinateAns))
	}else{
		if len(this.notSameTopic)!=0{
			ret += fmt.Sprintf("may not select: %s", fileUtil.UintsListToString(this.notSameTopic))
		}
	}
	return ret
}

func (this *RecognitionResult) Print()  {
	if this.directAnsFnd{
		fmt.Println("find direct answer: ", this.directAns)
	}else if this.sameTagFnd{
		fmt.Println("find same tag: ", this.sameTagAns)
	}else if this.coordinateFnd{
		fmt.Println("find coordinate: ", this.coordinateAns)
	}else{
		fmt.Println("not find, nut may not select: ", this.notSameTopic)
	}
}

func TestImgRecognation()  {
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("input img ident: ")
		var input string
		fmt.Fscan(stdin, &input)
		imgIdent := parseToImgIdent(input, "_")

		SaveAImg(imgIdent[0], imgIdent[1:], "E:/recogition/")
		
		res := ImgRecognitionByImgIdent(imgIdent)
		res.Print()
	}
}

func ImgRecognitionByImgIdent(imgIdent []byte) *RecognitionResult {
	srcData := PickImgDB(imgIdent[0]).ReadFor(imgIdent[1:])
	return ImgRecognitionByImgData(srcData)
}

func ImgRecognitionByImgData(imgSrcData []byte) *RecognitionResult{
	clipIndexMap := ImgIndex.GetClipsIndexesFromImgSrcData(imgSrcData)
	clipIndexes := make([][]byte, config.CLIP_COUNTS_OF_IMG)
	for i:=uint8(0);i < uint8(config.CLIP_COUNTS_OF_IMG);i ++{
		clipIndexes[int(i)] = (*clipIndexMap)[i]
	}

	imgIndex := ImgIndex.GetImgIndexByClipIndexes(clipIndexes)
	if len(imgIndex) != ImgIndex.IMG_INDEX_BYTES_LEN_int{
		fmt.Println("img index length is not ", ImgIndex.IMG_INDEX_BYTES_LEN_int, " : ", len(imgIndex))
		return  nil
	}

	ret := RecognitionResult{}

	//查找答案库
	directAns := InitTrainImgAnswerDB().ReadFor(imgIndex)
	if 0 != len(directAns){
		ret.directAnsFnd = true
		ret.directAns = directAns
		return &ret
	}

	//查找 clip 标签库
	tagToWhich := imgCache.NewMyMap(true)
	for i,clipIndex := range clipIndexes{
		curTag := QueryTagByClipIndex(clipIndex)
		if len(curTag) == 0{
			continue
		}
		tagToWhich.Put(curTag, i)
	}
	tags := tagToWhich.KeySet()

	for _,tag := range tags{
		interfaceWhichs := tagToWhich.Get(tag)
		if len(interfaceWhichs) <= 1{
			continue
		}

		ret.sameTagFnd = true

		group := make([]uint8, len(interfaceWhichs))
		for i,which := range interfaceWhichs{
			group[i] = uint8(which.(int))
		}

		ret.sameTagAns = append(ret.sameTagAns, group)
	}

	//找到了相同标签的子图, 此时不应返回, 应该继续找协同关系子图
	//if ret.sameTagFnd{
	//	return &ret
	//}


	//从协同库查找, 且使用 not same topic 过滤
	supportThreshold := 2
	coordinateResults := GetCoordinateClipsInClipIndexes(clipIndexes, supportThreshold)
	if len(coordinateResults) > 0{
		sort.Sort(coordinateResults)
		for _,coordinateRes := range coordinateResults{
			leftClipIndex := clipIndexes[coordinateRes.Left]
			rigthClipIndex := clipIndexes[coordinateRes.Right]
			if IsNotSameTopicOfClipIndexes(leftClipIndex, rigthClipIndex){
				fmt.Println("hit in not same top: ", coordinateRes.Left, "|", coordinateRes.Right)
				continue
			}else{
				ret.coordinateFnd = true

				ret.coordinateAns = append(ret.coordinateAns, []uint8{coordinateRes.Left, coordinateRes.Right})
			}
		}
	}
	if ret.coordinateFnd{
		return &ret
	}

	//标注非相同主题的子图, 可用于选择指导
	for i, clipIndex := range clipIndexes{
		for j:=i+1;j < int(config.CLIP_COUNTS_OF_IMG);j ++{
			if IsNotSameTopicOfClipIndexes(clipIndex, clipIndexes[j]){
				ret.notSameTopic = append(ret.notSameTopic, []uint8{uint8(i), uint8(j)})
			}
		}
	}

	return &ret
}


func parseToImgIdent(input string, splitStr string) []byte {
	if len(input) != 10{
		return  nil
	}
	ret := make([]byte, ImgIndex.IMG_IDENT_LENGTH)

	groups := strings.Split(input, splitStr)
	dbId,_ := strconv.Atoi(groups[0])

	imgKey := ImgIndex.FormatImgKey([]byte(groups[1]))

	ret[0] = uint8(dbId)
	copy(ret[1:], imgKey)

	return ret
}
