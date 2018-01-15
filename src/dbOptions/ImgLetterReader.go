package dbOptions

import "fmt"

func RandomReadOne()  {
	imgIndexDB := InitIndexToImgDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db failed")
		return
	}



	//imgIndexDB.DBPtr.Get
}