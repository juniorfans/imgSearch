package main

import (
	"bufio"
	"os"
	"fmt"
	"github.com/Comdex/imgo"
	"dbOptions"
)

func main(){
	dbOptions.DownloaderRun()
}



func  TestRead()  {
	stdin := bufio.NewReader(os.Stdin)
	baseDir := "E:/gen/3/"
	var input string
	for {
		fmt.Println("input image name to read")
		fmt.Fscan(stdin, &input)

		_ , err := imgo.Read(baseDir + input + ".jpg")
		if err != nil{
			fmt.Println("open jpg error : ", err)
		}else{
			fmt.Println("open success")
		}
	}


}



