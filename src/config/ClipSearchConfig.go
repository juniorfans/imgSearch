package config

import (
	"strings"
	"strconv"
	"io"
	"util"
	"fmt"
	"os"
	"bufio"
)

type ClipSearchConf struct {
	Delta_sd float64
	Delta_mean float64
	Delta_Eul float64
}

func (this* ClipSearchConf) Print()  {
	fmt.Println("Delta_sd: ", this.Delta_sd, ", Delta_mean: ", this.Delta_mean, ", Delta_Eul: ", this.Delta_Eul)
}

var clipSearchConf *ClipSearchConf = nil

func MustReReadSearchConf(confpath string)  {
	clipSearchConf = nil
	ReadClipSearchConf(confpath)
}

func ReadClipSearchConf(confPath string) *ClipSearchConf{

	if nil != clipSearchConf{
		return clipSearchConf
	}

	ret := ClipSearchConf{}
	clipSearchConf = &ret

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
			break
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
			if 0==strings.Compare("Delta_sd",kv[0]){
				ret.Delta_sd, _ = strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
			}else if 0 == strings.Compare("Delta_mean", kv[0]){
				ret.Delta_mean,_ = strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
			}else if 0 == strings.Compare("Delta_Eul", kv[0]){
				ret.Delta_Eul,_ = strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
			}else{
				fmt.Println("unknow db config key: ", line)
			}
		}
	}
	ret.Print()
	return &ret
}
