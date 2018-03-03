package dbOptions

import "imgSearch/src/imgCache"

//从多个表里面读取 key- value
type MultyDBReader struct {
	readRes chan []byte
	dbs []*DBConfig

	queryCached *imgCache.MyMap
}

func NewMultyDBReader(dbs []*DBConfig, cacheUse bool) *MultyDBReader {
	if 0 == len(dbs){
		return nil
	}
	ret := &MultyDBReader{}
	ret.dbs = dbs
	ret.readRes = make(chan []byte, len(dbs))
	if cacheUse{
		ret.queryCached = imgCache.NewMyMap(false)
	}else{
		ret.queryCached = nil
	}

	return ret
}

func (this *MultyDBReader) Close()  {
	close(this.readRes)
}


//read------------------------------------------------------------------------------
func (this *MultyDBReader)ReadFor(key []byte) [][]byte {
	if nil != this.queryCached{
		if this.queryCached.Contains(key){
			interfaceHits := this.queryCached.Get(key)
			if 0 == len(interfaceHits){
				return nil
			}
			return interfaceHits[0].([][]byte)

		}else{
			value := this.readDBs(key)
			this.queryCached.Put(key, value)
			return value
		}
	}

	return this.readDBs(key)
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
