package main

import (
	"dbOptions"
	"fmt"
	"time"
	"math/rand"
	"config"
	"strconv"
	"strings"
)

func main()  {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i:=0;i < 10;i++ {
		seqNo := r.Intn(9999999)
		letter:=r.Intn(8)
		imgId := string(dbOptions.MakeSurePlainImgIdIsOk([]byte(string(config.ThreadIdToName[letter]) + strconv.Itoa(seqNo))))
		newImgId := (string(dbOptions.ParseImgKeyToPlainTxt(dbOptions.FormatImgKey([]byte(imgId)))))
		if 0 != strings.Compare(imgId, newImgId){
			fmt.Println("error, imgId: ", imgId, ", newImgId: ", newImgId)
		}

	}

}
