package dbOptions

import (
	"util"
	"fmt"
	"os"
	"bufio"
	"io"
	"strings"
	"strconv"
)

type DBInitParams struct {
	DirBase string			//db dir
	BlockSize int			//sst 文件中块的字节数
	CompactionTableSize int		//compaction 生成的 sst 文件字节数
	BlockCacheCapacity int		//全局唯一的 block cache 的字节数 : 用于读
	WriteBuffer int			//写缓存大小. 若当前的 memtable 的大小已经到达了此大小则: 1)若 imm 不存在则等待将它的 compaction 完成
					//2) 将 mem dump 为 imm，再年是否需要 compaction
	CompactionL0Trigger int		//L0 文件个数达到这个阈值则触发 compaction
	CompactionTotalSize int		//第 i 层的 sst 文件的基数大小. 第 i 层大小是 CompactionTotalSize*(CompactionTotalSizeMultiplier^i).
					//CompactionTotalSizeMultiplier 默认值是 10M. 则第 1 层是 100 M, 第 2 层是 1000M

	WriteL0PauseTrigger int		//当 L0 的文件个数达到阈值则转暂停写入, 等待当前的 compact memtable 完成
	WriteL0SlowdownTrigger int	//当 L0 的文件个数达到阈值则转放慢写入的速度, sleep
}

func (this *DBInitParams) PrintLn ()  {
	fmt.Println(
		"dirBase: ", this.DirBase,
		", BlockSize: ", this.BlockSize/1024, " KB",
		", CompactionTableSize: ", this.CompactionTableSize/1024/1024, "MB",
		", BlockCacheCapacity: ", this.BlockCacheCapacity/1024/1024, "MB",
		", WriteBuffer: ", this.WriteBuffer/1024/1024,"MB",
		", CompactionL0Trigger: ", this.CompactionL0Trigger,
		", CompactionL0TotalSize: ", this.CompactionTotalSize/1024/1024,"MB",
		", WriteL0PauseTrigger: ", this.WriteL0PauseTrigger/1024/1024,"MB",
		", WriteL0SlowdownTrigger: ", this.WriteL0SlowdownTrigger/1024/1024,"MB")
}


func ReadDBConf(confPath string) *DBInitParams{

	ret := DBInitParams{}

	exedir,err := fileUtil.GetCurrentMoudlePath()
	if nil != err{
		fmt.Println("get current moudle error: ", err)
		return nil
	}
	f, err := os.Open(exedir + "/" + confPath)
	if err != nil {
		fmt.Println("open config error ", confPath, " : ", err)
		return nil
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')

		if io.EOF == err{
			return &ret
		}else if nil != err {
			fmt.Println("read config error")
			return nil
		}

		if 0 == strings.Index(line, "#") || 0 == strings.Index(line, "//"){
			continue
		}

		line = strings.TrimSpace(line)
		if 0 == len(line){
			continue
		}
		kv := strings.Split(line,"=")
		if 2 == len(kv){
			if 0==strings.Compare("DirBase",kv[0]){
				ret.DirBase = strings.TrimSpace(kv[1])
			}else if 0 == strings.Compare("BlockSize.bytes", kv[0]){
				ret.BlockSize,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("CompactionTableSize.bytes", kv[0]){
				ret.CompactionTableSize,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("BlockCacheCapacity.bytes", kv[0]){
				ret.BlockCacheCapacity,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("WriteBuffer.bytes", kv[0]){
				ret.WriteBuffer,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("CompactionL0Trigger.counts", kv[0]){
				ret.CompactionL0Trigger,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("CompactionTotalSize.bytes", kv[0]){
				ret.CompactionTotalSize,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("WriteL0PauseTrigger.bytes", kv[0]){
				ret.WriteL0PauseTrigger,_ = strconv.Atoi(kv[1])
			}else if 0 == strings.Compare("WriteL0SlowdownTrigger.bytes", kv[0]){
				ret.WriteL0SlowdownTrigger,_ = strconv.Atoi(kv[1])
			}else{
				fmt.Println("unknow db config key: ", line)
			}
		}
	}
	fmt.Println("to return db init params")
	ret.PrintLn()
	return &ret
}