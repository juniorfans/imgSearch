package dbOptions

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"bufio"
	"strings"
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

func SaveMainImg(mainImgKey ,dir string)  {
	imgDb := InitImgDB()
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