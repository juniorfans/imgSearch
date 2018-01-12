package dbOptions

import "fmt"

func RandomReadOne()  {
	imgIndexDB := InitImgIndexDB()
	if nil == imgIndexDB{
		fmt.Println("open img index db failed")
		return
	}



	//imgIndexDB.DBPtr.Get
}