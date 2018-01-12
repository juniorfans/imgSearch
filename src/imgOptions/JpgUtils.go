package ImgOptions

import (
	"io"
	"bytes"
	"image/jpeg"
	"fmt"
	"github.com/Comdex/imgo"
)


/**
	从 image 二进制字节到 image 结构化像素数据转化
 */
func FromImageFlatBytesToStructBytes(srcData []byte) [][][]uint8 {
	var reader io.Reader = bytes.NewReader([]byte(srcData))

	image, err := jpeg.Decode(reader)
	if nil != err{
		srcData = FixImage(srcData)
		reader = bytes.NewReader([]byte(srcData))
		image, err = jpeg.Decode(reader)
		if nil != err{
			fmt.Println("invalid image data: ", err)
			return nil
		}
	}
	data, err := imgo.Read(image)
	if nil != err{
		fmt.Println("read jpeg data error: ", err)
		return nil
	}
	return data
}
