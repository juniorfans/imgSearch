package ImgOptions

import (
	"github.com/Comdex/imgo"
	"fmt"
	"strings"
	"util"
)

func MergeEachDirInDir(path string, skipIfFileHas []string, suffixToAdd string, quality int)  {
	dirs := fileUtil.GetDirInDir(path)
	if(0 == len(dirs)){
		fmt.Println("path is empty: ", path)
		return
	}
	for _,dir := range dirs{
		MergeJpgIndir(path + "/" + dir, skipIfFileHas, suffixToAdd,quality)
	}
}

func MergeJpgIndir(path string, skipIfFileHas []string, suffixToAdd string, quality int)  {
	files := fileUtil.GetFilesInDir(path)
	if(len(files) == 0){
		fmt.Println("dir ", path," is empty")
		return
	}

	MergeJpgs(path,files,skipIfFileHas, suffixToAdd, quality)
	fmt.Println("merge jpg in path finished: ", path)
}

func MyImageFusion(imgMatrixLeft , imgMatrixRight [][][]uint8)(imgMatrix [][][]uint8) {

	height:=len(imgMatrixLeft)
	width:=len(imgMatrixLeft[0])

	for i:=0;i<height;i++{
		for j:=0;j<width;j++{
			imgMatrixLeft[i][j][0] = uint8(float64(imgMatrixLeft[i][j][0])*0.5)+uint8(float64(imgMatrixRight[i][j][0])*0.5)
			imgMatrixLeft[i][j][1] = uint8(float64(imgMatrixLeft[i][j][1])*0.5)+uint8(float64(imgMatrixRight[i][j][1])*0.5)
			imgMatrixLeft[i][j][2] = uint8(float64(imgMatrixLeft[i][j][2])*0.5)+uint8(float64(imgMatrixLeft[i][j][2])*0.5)
		}
	}
	imgMatrix = imgMatrixLeft
	return
}


func MyImageDiff(imgMatrixLeft , imgMatrixRight [][][]uint8)(imgMatrix [][][]uint8) {

	height:=len(imgMatrixLeft)
	width:=len(imgMatrixLeft[0])

	for i:=0;i<height;i++{
		for j:=0;j<width;j++{
			r1 := imgMatrixLeft[i][j][0]
			r2 := imgMatrixRight[i][j][0]
			var rdiff uint8
			if r1 == r2{
				rdiff =0
			}else if(r1 < r2){
				rdiff = r2-r1
			}else{
				rdiff = r1-r2
			}

			g1 := imgMatrixLeft[i][j][1]
			g2 := imgMatrixRight[i][j][1]
			var gdiff uint8
			if g1==g2{
				gdiff = 0
			}else if g1 < g2{
				gdiff = g2-g1
			}else{
				gdiff = g1-g2
			}

			b1 := imgMatrixLeft[i][j][2]
			b2 := imgMatrixRight[i][j][2]
			var bdiff uint8
			if b1==b2{
				bdiff = 0
			}else if b1 < b2{
				bdiff = b2-b1
			}else{
				bdiff = b1-b2
			}

			fmt.Println(rdiff, " ", gdiff, " ", bdiff," r1=", r1,", r2=",r2," g1=", g1,", g2=",g2, " b1=", b1,", b2=",b2)
			imgMatrixLeft[i][j][0] = 255-rdiff
			imgMatrixLeft[i][j][1] = 255-gdiff
			imgMatrixLeft[i][j][2] = 255-bdiff
		}
	}
	imgMatrix = imgMatrixLeft
	return
}

func MyImageFusion01(imgMatrixLeft , imgMatrixRight [][][]uint8)(imgMatrix [][][]uint8) {

	height:=len(imgMatrixLeft)
	width:=len(imgMatrixLeft[0])

	for i:=0;i<height;i++{
		for j:=0;j<width;j++{
			imgMatrixLeft[i][j][0] = imgMatrixLeft[i][j][0] | imgMatrixRight[i][j][0]
			imgMatrixLeft[i][j][1] = imgMatrixLeft[i][j][1] | imgMatrixRight[i][j][1]
			imgMatrixLeft[i][j][2] = imgMatrixLeft[i][j][2] | imgMatrixRight[i][j][2]
		}
	}
	imgMatrix = imgMatrixLeft
	return
}

func MergeJpgs(path string, files [] string, skipIfFileNameHas []string, suffixToAdd string, quality int)  {
	if 0 == len(files){
		return
	}

	//找到第一个不含 skipIfFileNamehas 的文件
	var start int=0
	for ;start < len(files);start++{
		skip := false
		if len(skipIfFileNameHas)!=0{
			for _,toSkip := range skipIfFileNameHas{
				if strings.Contains(files[start], toSkip){
					skip = true
					break
				}
			}
		}
		if !skip{
			break
		}
	}

	imgMatrixLeft,err := imgo.Read(path + "/" + files[start])

	if nil != err{
		fmt.Println("open file error: ", path + "/" + files[0], "merge error. ", err)
		return
	}
	height:=len(imgMatrixLeft)
	width:=len(imgMatrixLeft[0])
	for i,file := range files{
		skip := false
		if len(skipIfFileNameHas)!=0{
			for _,toSkip := range skipIfFileNameHas{
				if strings.Contains(file, toSkip){
					skip = true
					break
				}
			}
		}
		if skip{
			continue
		}

		if 0 == i{
			continue
		}else{
			imgMatrixRight,err := imgo.ResizeForMatrix(path + "/" + file,width,height)
			if nil != err{
				fmt.Println("resize file error: ", path + "/" + file, "merge error. ", err)
				return
			}
			imgMatrixLeft = MyImageFusion01(imgMatrixLeft,imgMatrixRight)
		}
	}

	newName := strings.Replace(path + "/" + files[0], ".", suffixToAdd + ".", 1)

	err = imgo.SaveAsJPEG(newName, imgMatrixLeft, quality)
	if nil == err{
		fmt.Println("merge success")
	}else{
		fmt.Println("merge failed, save jpg error. ", err)
	}
}
