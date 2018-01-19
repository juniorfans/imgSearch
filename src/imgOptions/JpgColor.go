package ImgOptions

import "image/color"

func TranRGBToYCBCR(r , g , b uint8) (y, cb, cr uint8 ){
	y, cb, cr = color.RGBToYCbCr(r, g, b)
	return
}