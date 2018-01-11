package ImgOptions

import (
	"util"
	"fmt"
	"strings"
	"github.com/Comdex/imgo"
)

func MyBinaryzation(src [][][]uint8, threshold int) [][][]uint8 {
	imgMatrix := imgo.RGB2Gray(src)

	height:=len(imgMatrix)
	width:=len(imgMatrix[0])
	for i:=0; i<height; i++ {
		for j:=0; j<width; j++ {
			var rgb int = int(imgMatrix[i][j][0])+int(imgMatrix[i][j][1])+int(imgMatrix[i][j][2])
			//fmt.Println(rgb)
			if rgb > threshold {
				rgb = 255
			}else{
				rgb = 0
			}
			imgMatrix[i][j][0]=uint8(rgb)
			imgMatrix[i][j][1]=uint8(rgb)
			imgMatrix[i][j][2]=uint8(rgb)
		}
	}

	return imgMatrix
}

func JpgInDirBinarayzation(path string, skipIfFileNameHas []string, suffixToAdd string ,threshold int)  {
	files := fileUtil.GetFilesInDir(path)
	if len(files)==0{
		fmt.Println("path is empty: ", path)
		return
	}
	for _,file := range files{
		skip := false
		if len(skipIfFileNameHas)!=0{
			for _,toSkip := range skipIfFileNameHas{
				if strings.Contains(file, toSkip){
					skip = true
					break
				}
			}
		}
		if !skip{
			JpgBinaryzation(path+"/"+file, suffixToAdd, threshold)
		}
	}
}

func DirInDirBinarayzation(path string, skipIfFileNameHas []string, suffixToAdd string, threshold int) {
	dirs := fileUtil.GetDirInDir(path)
	if len(dirs) == 0{
		fmt.Println("path contains no dir: ", path)
		return
	}
	for _,dir := range dirs{
		JpgInDirBinarayzation(path + "/" + dir, skipIfFileNameHas, suffixToAdd, threshold)
	}
}

func JpgBinaryzation(fileName , suffixToAdd string, threshold int)  {
	imgMatrix:=imgo.MustRead(fileName)
	imgMatrix_gray := MyBinaryzation(imgMatrix, threshold)
	gray_rgb := strings.Replace(fileName, ".", suffixToAdd + ".", 1)
	err := imgo.SaveAsJPEG(gray_rgb, imgMatrix_gray, 100)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s generate finish\n", gray_rgb)
}
