package dbOptions

import (
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"config"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
)


type howManyCalInfo struct {
	threadId int
	count int
}

var howManyCalFinished chan howManyCalInfo

func howManyImagesForThread(imgDB *DBConfig, threadId int) {
	region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
	iter := imgDB.DBPtr.NewIterator(&region,&imgDB.ReadOptions)
	iter.First()
	fmt.Println("thread: ", threadId, ", begin: ", string(iter.Key()))
	ncount := 0
	for iter.Valid(){
		ncount ++
		iter.Next()
	}

	howManyCalFinished <- howManyCalInfo{threadId:threadId, count:ncount}
}

func ImgDBStatRepair(imgDB *DBConfig) (count int, realMaxCores uint8) {
	return CalcAndSaveImageCounts(imgDB)
}

func CalcAndSaveImageCounts(imgDB *DBConfig) (count int, realMaxCore uint8)  {
	maxCore := config.MAX_THREAD_COUNT
	realMaxCore = 0

	howManyCalFinished = make(chan howManyCalInfo, maxCore)
	for i:=0;i < maxCore ;i++  {
		go howManyImagesForThread(imgDB, i)
	}

	count = 0
	for i:=0;i < maxCore ;i++  {
		calInfo := <- howManyCalFinished
		if 0 != calInfo.count{
			fmt.Println("threadId appear: ", calInfo.threadId)
			realMaxCore++
		}
		count += calInfo.count
	}
	fmt.Println("img count: ", count, ", real max cores: ", realMaxCore)
	imgDB.WriteTo(config.STAT_KEY_DOWNLOAD_MAX_CORES, []byte(strconv.Itoa(int(realMaxCore))))
	imgDB.WriteTo(config.STAT_KEY_DOWNLOAD_BASE, []byte(strconv.Itoa(count)))
	return
}

func HowManyImages() int  {
	imgDB := initImgDB()
	if nil == imgDB{
		return 0
	}
	countStr := string(imgDB.ReadFor(config.STAT_KEY_DOWNLOAD_BASE))
	ret, _ := strconv.Atoi(countStr)
	fmt.Println("total size: ", ret)
	return ret
}

func HowManyImageClipIndexes(dbId uint8) int  {
	return RepairTotalSize(InitMuIndexToClipDB(dbId))
}


func setStatInfo(imgDB *DBConfig, lastBase int, cores int, eachTimes int, costSecs int64)  {
	db := imgDB.DBPtr

	maxCores := cores
	maxCoreStr := imgDB.ReadFor(config.STAT_KEY_DOWNLOAD_MAX_CORES)
	if nil!= maxCoreStr{
		mc, err := strconv.Atoi(string(maxCoreStr))
		if err==nil{
			if mc > maxCores{
				maxCores = mc
			}
		}
	}

	writeOptions := &imgDB.WriteOptions

	db.Put(config.STAT_KEY_DOWNLOAD_BASE, []byte(strconv.Itoa(lastBase)), writeOptions)

	db.Put(config.STAT_KEY_DOWNLOAD_CORES, []byte(strconv.Itoa(cores)), writeOptions)

	db.Put(config.STAT_KEY_DOWNLOAD_MAX_CORES, []byte(strconv.Itoa(maxCores)), writeOptions)

	db.Put(config.STAT_KEY_DOWNLOAD_EACH_TIMES, []byte(strconv.Itoa(eachTimes)), writeOptions)

	db.Put(config.STAT_KEY_DOWNLOAD_COST_SECS, []byte(strconv.FormatInt(costSecs, 10)), writeOptions)

	key := string(config.STAT_KEY_DOWNLOAD_STAT) + time.Now().Format("2006-01-02 15:04:05");
	value := "base: " + strconv.Itoa(lastBase) + ", cores:" + strconv.Itoa(cores) + ", eachTimes:" + strconv.Itoa(eachTimes) + ", cost:"+strconv.FormatInt(costSecs, 10)+"s"
	db.Put([]byte(key),[]byte(value), writeOptions)
	db.Put(config.STAT_KEY_DOWNLOAD_CUR_STAT_KEY, []byte(key),  writeOptions)
}

func GetStatInfo(dbConfig *DBConfig) (lastBase int, maxCores, lastCores int, lastEachTimes int, lastCostScs int64, remark string){

	db := dbConfig.DBPtr
	readOptions := &dbConfig.ReadOptions

	{
		cs,err := db.Get(config.STAT_KEY_DOWNLOAD_CUR_STAT_KEY, readOptions)
		if err==leveldb.ErrNotFound{
			remark = "no stat data"
		}else{
			ret, err:= db.Get([]byte(cs), readOptions)
			if err == leveldb.ErrNotFound{
				remark = "no stat data"
			}else{
				remark = string(ret)
			}
		}
	}


	{
		lb, err := db.Get(config.STAT_KEY_DOWNLOAD_BASE, readOptions)
		if err == leveldb.ErrNotFound{
			lastBase = 0
		}else{
			ilb,err:=strconv.Atoi(string(lb))
			if err==nil{
				lastBase = ilb
			}else{
				lastBase =0
			}
		}
	}


	{
		lc, err := db.Get(config.STAT_KEY_DOWNLOAD_CORES, readOptions)
		if err == leveldb.ErrNotFound{
			lastCores = 0
		}else{
			ilc,err:=strconv.Atoi(string(lc))
			if err==nil{
				lastCores = ilc
			}else{
				lastCores =0
			}
		}
	}

	{
		mc, err := db.Get(config.STAT_KEY_DOWNLOAD_MAX_CORES, readOptions)
		if err == leveldb.ErrNotFound{
			maxCores = 0
		}else{
			imc,err:=strconv.Atoi(string(mc))
			if err==nil{
				maxCores = imc
			}else{
				maxCores =0
			}
		}
	}

	{
		let, err := db.Get(config.STAT_KEY_DOWNLOAD_EACH_TIMES, readOptions)
		if err == leveldb.ErrNotFound{
			lastEachTimes = 0
		}else{
			ilet,err:=strconv.Atoi(string(let))
			if err==nil{
				lastEachTimes = ilet
			}else{
				lastEachTimes =0
			}
		}
	}

	{
		lcost, err := db.Get(config.STAT_KEY_DOWNLOAD_COST_SECS, readOptions)
		if err == leveldb.ErrNotFound{
			lastCostScs = 0
		}else{
			icost,err:=strconv.ParseInt(string(lcost),10,64)
			if err==nil{
				lastCostScs = icost
			}else{
				lastCostScs =0
			}
		}
	}
	return
}
