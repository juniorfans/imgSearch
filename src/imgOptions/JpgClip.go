package ImgOptions

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"strings"
	"path/filepath"

)

func ImgClipInDir(path string){

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		fileName := path//f.Name()
		if strings.Contains(fileName,"_small"){
			return nil;
		}

		ce := ImgClip(fileName, 50)
		if nil != ce{

		}
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}


func ImgClip(fileName string, quality int) error {
	src := fileName
	dst := strings.Replace(src, ".", "_small.", 1)
	fmt.Println("src=", src, " dst=", dst)
	fIn, _ := os.Open(src)
	defer fIn.Close()

	fOut, _ := os.Create(dst)
	defer fOut.Close()

	err := clip(fIn, fOut, 120, 0, 180, 25, quality)
	if err != nil {
		fmt.Println("file clip error: ", fileName)
		return err
	}
	return nil
}

func ImgClipByPoints(fileName string,x0, y0, x1, y1, quality int) error {
	src := fileName
	dst := strings.Replace(src, ".", "_small.", 1)
	fmt.Println("src=", src, " dst=", dst)
	fIn, _ := os.Open(src)
	defer fIn.Close()

	fOut, _ := os.Create(dst)
	defer fOut.Close()

	err := clip(fIn, fOut, x0, y0, x1, y1, quality)
	if err != nil {
		fmt.Println("file clip error: ", fileName)
		return err
	}
	return nil
}


/*
* 图片裁剪
* 入参:
* 规则:如果精度为0则精度保持不变
*
* 返回:error
 */
func clip(in io.Reader, out io.Writer, x0, y0, x1, y1, quality int) error {
	origin, fm, err := image.Decode(in)
	if err != nil {
		return err
	}

	switch fm {
	case "jpeg":
		img := origin.(*image.YCbCr)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)
		return jpeg.Encode(out, subImg, &jpeg.Options{quality})
	default:
		return errors.New("ERROR FORMAT")
	}
	return nil
}

