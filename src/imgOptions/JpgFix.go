package ImgOptions

import (
	"fmt"
	"io/ioutil"
	"strings"
	"bufio"
	"os"
)

/**

	jpeg 图片格式解析异常，经过查找定位，有以下两个原因：

	1)非法的 markder，根据 image.jpeg 库的 reader.go 中描述的 markder，无效的 marker 均小于 C0
	2)Corrupt JPEG data: 'n' extraneous bytes before marker 0xd9 ，意思是说marker: 0xd9 的前面有 n 个多余的字节

	经过排查，发现 1) 会导致解析错误，而 2) 只是一个告警，不会报错。所以我们的解决办法即是删除不合法的 marker 的第一个字节：0xFF

	补上一点 jpeg 文件格式 (marker)：
	https://gist.github.com/zwn/5d4e7cdef308d6a8eb8e5f4da19523d7
 */

func mergeBytes(s1 []byte, s2 []byte) []byte {
	slice := make([]byte, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}

func FixImageByName(fileName string)  {
	srcData ,err := ioutil.ReadFile(fileName)
	if nil != err{
		fmt.Println("open file error, ", fileName, err)
		return
	}

	newData := FixImage(srcData)

	fixed := strings.Replace(fileName, ".", "_fixed_.", 1)
	ioutil.WriteFile(fixed,newData,0644)
}

func FixImage(srcData []byte) []byte {
	newData := DeleteInvalidMarker(srcData)
	return newData
}


func TestJpgFix()  {
	stdin := bufio.NewReader(os.Stdin)
	baseDir := "E:/gen/1/"
	var input string
	var marker byte
	var nsize int
	for{
		fmt.Println("input image keys to fix, marker and nsize")
		fmt.Fscan(stdin,&input, &marker, &nsize)
		PrintBytesAroundMarkder(baseDir + input + ".jpg", marker, nsize)
		//ImgOptions.DeleteInvalidMarker(baseDir + input + ".jpg", marker)
		//		ImgOptions.DeleteBytesBeforeMarker(baseDir + input + ".jpg", marker, nsize)
		//	ImgOptions.FixImageByName(baseDir + input + ".jpg")
	}
}

/**
	如果 0xFF 后面一个字节小于 D0 则认为是一个错误的 marker 则需要将 0xFF 删除

 */
func DeleteInvalidMarker(srcData []byte) []byte {
	nsize := len(srcData)
	ret := make([]byte, nsize)
	ic := 0
	for i:=0;i < nsize;{
		if srcData[i] != 0xff{
			ret[ic] = srcData[i]
			ic ++
			i++
		}else{
			if i+1 < nsize && 0x00 != srcData[i+1] && srcData[i+1] < 0xC0{
				//find a invalid marker
			//	fmt.Printf("%X ", srcData[i+1])
				ret[ic] = srcData[i+1]
				ic ++
			}else{
				ret[ic] = srcData[i]
				ic ++
				ret[ic] = srcData[i+1]
				ic ++
			}
			i += 2
		}
	}

//	fmt.Println("delete ", nsize-ic, " invalid marker")

	return ret[0:ic]
}

func PrintBytesAroundMarkder(fileName string, marker byte, arroundSize int)  {

	srcData ,err := ioutil.ReadFile(fileName)
	if nil != err{
		fmt.Println("open file error, ", fileName, err)
		return
	}

	pos := GetMarkerPos(srcData, marker)
	if -1 == pos{
		fmt.Printf("can't find markder: %X \n", marker)
		return
	}
	for i:=pos-arroundSize ;i < pos;i ++{
		fmt.Printf("%X ", srcData[i])
	}
	fmt.Println("\n---------------")
	for i:=pos ;i < len(srcData) && i < pos+arroundSize;i ++{
		fmt.Printf("%X ", srcData[i])
	}
	fmt.Println()

}

func GetMarkerPos(srcData []byte, marker byte) int {
	nsize := len(srcData)
	pos := -1
	for i:=0;i < nsize;{
		if srcData[i] != 0xff{
			i++
		}else{
			if i+1 < nsize && srcData[i+1] == marker{
				pos = i //marker 的位置(含 ff)
				break
			}else{
				i += 2
			}
		}
	}
	return pos
}
