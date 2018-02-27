package config

import (

)


type ClipSearchConf struct {
	Delta_sd float64
	Delta_mean float64
	Delta_Eul_square float64
	Delta_Eul float64
}

/*

func (this* ClipSearchConf) Print()  {
	fmt.Println("Delta_sd: ", this.Delta_sd, ", Delta_mean: ", this.Delta_mean, ", Delta_Eul_square: ", this.Delta_Eul_square)
}

var TheclipSearchConf *ClipSearchConf = nil

func MustReReadSearchConf(confpath string)  {
	TheclipSearchConf = nil
	ReadClipSearchConf(confpath)
}

func ReadClipSearchConf(confPath string) *ClipSearchConf{

	if nil != TheclipSearchConf {
		return TheclipSearchConf
	}

	ret := ClipSearchConf{}
	TheclipSearchConf = &ret

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
			}else if 0 == strings.Compare("Delta_Eul_square", kv[0]){
				ret.Delta_Eul_square,_ = strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
			}else{
				fmt.Println("unknow db config key: ", line)
			}
		}
	}
	ret.Print()
	return &ret
}

*/