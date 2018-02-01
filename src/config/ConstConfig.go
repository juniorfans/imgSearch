package config

import "util"

//系统最多只能允许 25 个线程. 参考 ImgDBKeyFormatDump : region := util.Range{Start:[]byte{config.ThreadIdToByte[threadId]}, Limit:[]byte{config.ThreadIdToByte[threadId+1]}}
var MAX_THREAD_COUNT = 25

var ThreadIdToName = map[int]string{
	0:"A",1:"B",2:"C",3:"D",4:"E",5:"F",6:"G",7:"H",
	8:"I",9:"J",10:"K",11:"L",12:"M",13:"N",14:"O",15:"P",
	16:"Q",17:"R",18:"S",19:"T",20:"U",21:"V",22:"W",23:"X",
	24:"Y",25:"Z",
}

var ThreadIdToByte = map[int]byte{
	0:'A',1:'B',2:'C',3:'D',4:'E',
	5:'F',6:'G',7:'H',8:'I',9:'J',
	10:'K',11:'L',12:'M',13:'N',14:'O',
	15:'P',16:'Q',17:'R',18:'S',19:'T',
	20:'U',21:'V',22:'W',23:'X',24:'Y',
	25:'Z',
}

var STAT_KEY_PREX []byte = []byte("_ZLAST")
var STAT_KEY_TOTALSIZE_PREX []byte = []byte("_ZLAST_TOTAL")
var STAT_KEY_SORT_BY_VALUE_SIZE_PREX []byte = []byte("_ZLAST_SORTED_BY_VALUE_SIZE")

var STAT_KEY_DOWNLOAD_BASE []byte = []byte("_ZLAST_D_BASE")
var STAT_KEY_DOWNLOAD_CORES []byte = []byte("_ZLAST_D_CORES")
var STAT_KEY_DOWNLOAD_MAX_CORES []byte = []byte("_ZLAST_D_MAX_CORES")
var STAT_KEY_DOWNLOAD_EACH_TIMES []byte = []byte("_ZLAST_D_EACH_TIMES")
var STAT_KEY_DOWNLOAD_COST_SECS []byte = []byte("_ZLAST_D_COST_SECS")
var STAT_KEY_DOWNLOAD_STAT []byte = []byte("_ZLAST_D_STAT")
var STAT_KEY_DOWNLOAD_CUR_STAT_KEY []byte = []byte("_ZLAST_D_CUR_STAT_KEY")

func IsValidUserDBKey(key []byte) bool {
	return !fileUtil.BytesStartWith(key ,STAT_KEY_PREX)
}