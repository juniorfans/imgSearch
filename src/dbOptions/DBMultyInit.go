package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"fmt"
	"strconv"
)

var imgDBs map[uint8]*DBConfig = make(map[uint8]*DBConfig)

func GetImgDBs() []*DBConfig {
	fmt.Println("picked img db : ", len(imgDBs))
	ret := make([]*DBConfig, len(imgDBs))
	i := 0
	for _,db := range imgDBs{
		ret[i] = db
		i ++
	}
	return ret
}

func PickImgDB(dbId uint8) *DBConfig {
	ret := imgDBs[dbId]
	if nil == ret{
		dbDir := "D:/img_db_" +  strconv.Itoa(int(dbId))+ "/image.db"

		fmt.Println("has pick this img db: ", dbDir)

		imgDBConfig := DBConfig{
			Dir : dbDir,
			DBPtr : nil,
			OpenOptions : opt.Options{ErrorIfMissing:false},
			ReadOptions : opt.ReadOptions{},
			WriteOptions : opt.WriteOptions{Sync:false},
			inited : false,
			Id : dbId,
			Name:"img db",
		}

		_, err :=  initDB(&imgDBConfig)
		if err != nil{
			fmt.Println("open img db error, ", err)
			return nil
		}
		imgDBs[dbId] = &imgDBConfig
		ret = &imgDBConfig
	}else{
		initDB(ret)	//防止 DBConfig 被关闭但没有移除掉, 现在需要复用，所以要重新初始化
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

func removeClosed()  {
	var aliveDBs []*DBConfig
	for _,db := range imgDBs{
		if false != db.inited{
			aliveDBs = append(aliveDBs,db)
		}
	}

	imgDBs = make(map[uint8]*DBConfig, len(aliveDBs))
	for _,db := range aliveDBs{
		imgDBs[db.Id]=db
	}
}