package ImgOptions

import (
	"path/filepath"
	"os"
	"strings"
	"fmt"
	"github.com/Comdex/imgo"
)

func JpgDirCosSim(path string)  {
	files := make(map[int]string)
	count := 0
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		fileName := path//f.Name()
		if !strings.Contains(fileName,"_small.jpg"){
			return nil;
		}

		files[count]=fileName
		count ++

		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	for i,_ := range files{
		if(0 == i){
			continue
		}else{
			left := files[i-1]
			right := files[i]
			fmt.Println(left," -- ", right, " : ", JpgCosSim(left, right))
		}
	}

}

func JpgCosSim(file1 string, file2 string) float64 {
	cossimi ,err := imgo.CosineSimilarity(file1,file2)
	if nil != err{
		fmt.Println("cossim error: ", file1, " -- ", file2)
		return 0.0
	}
	return cossimi

}
