package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"bufio"
	"strings"
	"time"
	"strconv"
	"math/rand"
	"config"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"util"
)

func writeToFile(content []byte, fileName string)  {
	image, err := os.Create(fileName)
	if err != nil {
		fmt.Println("create file failed:", fileName, err)
		return
	}

	defer image.Close()
	image.Write(content)
}

func SaveMainImgsIn(mainImgKeys []string, dir string)  {

	os.MkdirAll(dir, 0777)

	for _, mainImgKey := range mainImgKeys{
		SaveMainImg(string(FormatImgKey([]byte(mainImgKey))), dir)
	}
}

func SaveImgLetterIn(mainImgKeys []string, dir string)  {

	os.MkdirAll(dir, 0777)

	for _, mainImgKey := range mainImgKeys{
		SaveMainImg(mainImgKey, dir)
	}
}

func SaveMainImgs()  {
	for  {
		SaveTheInputImg()
	}
}

func SaveTheInputImg()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string

	fmt.Println("input image keys to save, split by - ")
	fmt.Fscan(stdin,&input)
	keys := strings.Split(input,"-")
	SaveMainImgsIn(keys, "E:/gen/2/")

}


func SaveDuplicatedMostImg()  {
	stdin := bufio.NewReader(os.Stdin)
	var input int
	fmt.Println("how many count to save duplicated img")
	fmt.Fscan(stdin,&input)

	ReadIndexSortInfo(input)

}

func DeleteStatImgClipsInfo()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId uint8

	fmt.Println("input db id: ")

	fmt.Fscan(stdin,&dbId)

	clipDB:= InitImgClipsDB()

	for i:=0;i < 8 ;i++  {
		//lastKey,count := GetThreadLastDealedKey(clipDB,dbId,i)
		//fmt.Println("lastKey: ", string(lastKey), ", count: ", count)
		SetThreadLastDealedKey(clipDB, dbId,i,[]byte{},0)
	}
	clipDB.CloseDB()

}

func StatImgClipsInfo()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId uint8

	fmt.Println("input db id: ")

	fmt.Fscan(stdin,&dbId)

	clipDB:= InitImgClipsDB()

	for i:=0;i < 8 ;i++  {
		lastKey,count := GetThreadLastDealedKey(clipDB,dbId,i)
		fmt.Println("lastKey: ", string(lastKey), ", count: ", count)
	}
	clipDB.CloseDB()

}

func StatImgIndexesInfo()  {
	stdin := bufio.NewReader(os.Stdin)
	var dbId uint8

	fmt.Println("input db id: ")

	fmt.Fscan(stdin,&dbId)

	indexDB:= InitIndexToImgDB()

	for i:=0;i < 8 ;i++  {
		lastKey,count := GetThreadLastDealedKey(indexDB,dbId,i)
		fmt.Println("lastKey: ", string(lastKey), ", count: ", count)
	}
	indexDB.CloseDB()

}


func PrintAllStatInfo()  {
	PickImgDB(1)
	PickImgDB(2)
	PickImgDB(4)
	imgDBs := GetImgDBs()
	for _,imgDB := range imgDBs{
		imgDB.PrintStat()
	}

	clipDB := InitImgClipsDB()
	clipDB.PrintStat()

	indexDB := InitIndexToImgDB()
	indexDB.PrintStat()

	for _,imgDB := range imgDBs{
		imgDB.CloseDB()
	}
}

func SaveMainImg(mainImgKey ,dir string)  {
	imgDb := GetImgDBWhichPicked()
	if nil == imgDb{
		fmt.Println("open img db failed")
		return
	}

	imgData, err := imgDb.DBPtr.Get([]byte(mainImgKey), &imgDb.ReadOptions)
	if leveldb.ErrNotFound == err{
		fmt.Println("can't find img: ", mainImgKey)
		return
	}

	fileName := dir + string(filepath.Separator) + mainImgKey + ".jpg"
	writeToFile(imgData, fileName)
	fmt.Println(fileName, " save success")
}

//保存图片上的文字
func SaveImgLetter(mainImgKey ,dir string)  {
	imgDb := GetImgDBWhichPicked()
	if nil == imgDb{
		fmt.Println("open img db failed")
		return
	}

	imgData, err := imgDb.DBPtr.Get([]byte(mainImgKey), &imgDb.ReadOptions)
	if leveldb.ErrNotFound == err{
		fmt.Println("can't find img: ", mainImgKey)
		return
	}

	fileName := dir + string(filepath.Separator) + mainImgKey + ".jpg"


	writeToFile(imgData, fileName)
	fmt.Println(fileName, " save success")
}

func GetRandomImgKey() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	index := r.Intn(1000)

	letter:=r.Intn(8)

	return config.ThreadIdToName[letter] + strconv.Itoa(index)
}

func ReadClipValues()  {
	stdin := bufio.NewReader(os.Stdin)
	var input int

	fmt.Println("input how many count values for clip db- ")
	fmt.Fscan(stdin,&input)
	ReadClipValuesInCount(input)
}


func TestSaveAClipFromValues()  {
	InitImgClipsDB()
	stdin := bufio.NewReader(os.Stdin)
	var input int

	fmt.Println("input how many count values for clip db to save clips ")
	fmt.Fscan(stdin,&input)
	saveAClipFromValues(input)
	InitImgClipsDB().CloseDB()
}

func saveAClipFromValues(count int)  {
	iter := InitImgClipsDB().DBPtr.NewIterator(nil, &opt.ReadOptions{})

	if(!iter.First()){
		fmt.Println("seek to first error")
	}

	for iter.Valid(){
		//writeToFile(iter.Value(), string(iter.Key()))
		fmt.Println("-----------------------------------------------------")
		fileUtil.PrintBytes(iter.Key())
		SaveClipAsJpgFromIndexValue(iter.Value(), "E:/gen/cclip/")
		iter.Next()
		count --
		if count <= 0{
			break
		}
	}
	iter.First()
}

func SaveLetterOfImg()  {
	stdin := bufio.NewReader(os.Stdin)
	var input string

	fmt.Print("which img to save letter: ")
	fmt.Fscan(stdin,&input)

	SaveLetterIndexAsJpgForImg(GetImgDBWhichPicked(),[]byte(input), "E:/gen/letter/")
}