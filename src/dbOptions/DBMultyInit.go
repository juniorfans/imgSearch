package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"strconv"
)

var imgDBs map[int]*DBConfig = make(map[int]*DBConfig)

func PickImgDB(dbId int) *DBConfig {
	ret := imgDBs[dbId]
	if nil == ret{
		dbDir := "D:/img_db_" +  strconv.Itoa(dbId)+ "/image.db"

		fmt.Println("has pick this img db: ", dbDir)

		imgDBConfig := DBConfig{
			Dir : dbDir,
			DBPtr : nil,
			OpenOptions : opt.Options{ErrorIfMissing:false},
			ReadOptions : opt.ReadOptions{},
			WriteOptions : opt.WriteOptions{Sync:false},
			inited : false,
			Id : dbId,
		}

		_, err :=  initDB(&imgDBConfig)
		if err != nil{
			fmt.Println("open img db error, ", err)
			return nil
		}
		imgDBs[dbId] = &imgDBConfig
		ret = &imgDBConfig
	}else{
		initDB(ret)	//防止 DBConfig 被关闭但没有移除掉
	}

	return ret
}

func GetImgDBWhichPicked() *DBConfig {
	if 0 == len(imgDBs){
		fmt.Println("no picked img db")
		return nil
	}else if 1 == len(imgDBs){
		for _,ret := range imgDBs{
			return ret
		}
		fmt.Println("this must not be happen!")
		return nil
	}else{
		fmt.Println("too many picked img db")
		return nil
	}
}