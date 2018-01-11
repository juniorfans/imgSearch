package fileUtil

import (
	"os"
	"fmt"
	"io/ioutil"
)

func __main()  {
	image, err := os.OpenFile("E:/clip/H955_small.jpg", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("create file failed:", err)
		return
	}
	finfo,_ := image.Stat()

	var fsize int = int(finfo.Size())

	data := make([]byte, fsize)
	readSize , err := image.Read(data)

	if err != nil{
		fmt.Println("read file error:", err)
		return
	}
	if(readSize != fsize){
		fmt.Println("read file error:", err)
		return
	}

	for _, byte := range data{
		fmt.Print(",",byte)
	}
}


func GetDirInDir(path string) [] string  {
	dir, err := ioutil.ReadDir(path)
	if nil!= err{
		fmt.Println("open path erro,", err)
		return []string{}
	}

	files := make([]string,len(dir))

	for i,finfo := range dir{
		if finfo.IsDir() {
			files[i] = finfo.Name()
		}
	}

	return files
}

func GetFilesInDir(path string) []string {

	dir, err := ioutil.ReadDir(path)
	if nil!= err{
		fmt.Println("open path erro,", err)
		return []string{}
	}

	files := make([]string,len(dir))

	for i,finfo := range dir{
		if !finfo.IsDir() {
			files[i] = finfo.Name()
		}
	}

	return files
}

func WriteToFile(content []byte, img_dir , fileName string)  {
	image, err := os.Create(img_dir + "/" + fileName)
	if err != nil {
		fmt.Println("create file failed:", fileName, err)
		return
	}

	defer image.Close()
	image.Write(content)
}