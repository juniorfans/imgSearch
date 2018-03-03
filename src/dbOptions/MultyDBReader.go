package dbOptions

import (
	"imgCache"
)

//从多个表里面读取 key- value
type MultyDBReader struct {
	readRes chan []byte
	dbs []*DBConfig

	queryCached *imgCache.MyConcurrentMap
}


func NewMultyDBReader(dbs []*DBConfig, cached *imgCache.MyConcurrentMap) *MultyDBReader {
	if 0 == len(dbs){
		return nil
	}
	ret := &MultyDBReader{}
	ret.dbs = dbs
	ret.readRes = make(chan []byte, len(dbs))

	ret.queryCached = cached

	return ret
}

func (this *MultyDBReader) Close()  {
	close(this.readRes)
}


//read------------------------------------------------------------------------------
func (this *MultyDBReader)ReadFor(key []byte) (ret [][]byte , cacheHit bool){
	cacheHit = false
	if nil != this.queryCached{
		if this.queryCached.Contains(key){
			values := this.queryCached.Get(key)
			if len(values) != 1{
				return
			}

			ret = values[0].([][]byte)
			cacheHit = true
			return
		}else{
			value := this.readDBs(key)
			this.queryCached.Put(key, value)
			return
		}
	}else{
		ret = this.readDBs(key)
		return
	}
}

func (this *MultyDBReader)readDBs(key []byte) [][]byte {
	for i:=0;i < len(this.dbs);i ++{
		go this.readDB(this.dbs[i], key)
	}

	ret := make([][]byte, len(this.dbs))
	for i:=0; i < len(this.dbs); i++{
		r := <- this.readRes
		ret[i] = r
	}

	return ret
}


func (this *MultyDBReader)readDB(db *DBConfig, key []byte) {
	this.readRes <- db.ReadFor(key)
}
