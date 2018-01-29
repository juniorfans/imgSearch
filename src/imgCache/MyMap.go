package imgCache

import (
	"bytes"
	"fmt"
)

/**
	提供一个以 []byte 为键, []interface{} 为值的 map, 非线程安全
 */


/**
	myMapValue 作为 MyMap 的元素, 它保持了使用者传入的 key-value, 用于准确定位
 */
type myMapValue struct {
	values []interface{}
	key []byte
}

func (this *myMapValue) Clear()  {
	this.key = nil
	this.values = nil
}

type MyMap struct {
	//当 hash 碰撞时, 采用拉链法解决
	data map[int] []myMapValue

	multyValuesSupported bool

	//-------------------
	keyCount, valueCount int	//不同的 key, 不同的 value 的个数
}

func NewMyMap(multyValuesSupported bool) *MyMap {
	ret := MyMap{}
	ret.data = make(map[int] []myMapValue)
	ret.multyValuesSupported = multyValuesSupported
	ret.keyCount = 0
	ret.valueCount = 0
	return &ret
}

func (this *MyMap) GetKeyCounts() int{
	return this.keyCount
}

func (this *MyMap) GetValueCounts() int{
	return this.valueCount
}


func (this *MyMap) findKey(key []byte) (ret *myMapValue, hash int ){
	hash = this.getHashCode(key)

	ret = nil
	if hashConflicts, ok := this.data[hash]; ok{
		//当前 byte 的 hash 已存在槽, hash 存在碰撞, 碰撞意味着 hash 一样，但 key 可能一样/不一样的元素.
		//下面依次遍历该槽中所有的 myMapValue, 当找到了相同的 key 时, 则合并/替代

		//todo 此处可以优化为二分查找
		if 1<len(hashConflicts){
	//		fmt.Println("occ more :", len(hashConflicts))
		}
		for i,conflict := range hashConflicts {
			if bytes.Equal(conflict.key, key){
				ret = &hashConflicts[i]
				break
			}
		}
	}
	return
}

/**
	put key-value into the map. will not copy memory, only hold reference. so if change the slice outside,
	everything goes wrong: hashcode won't calc again
 */
func (this *MyMap)Put(key []byte, value interface{})  {
	whichMapValue, hash := this.findKey(key)

	if nil == whichMapValue{
		//key 不存在
		mapValue := myMapValue{key:key, values:[]interface{}{value}}
		if len(this.data[hash]) == 0{
			this.data[hash] = []myMapValue{mapValue}
		}else{
			this.data[hash] = append(this.data[hash], mapValue)
		}

		//更改计数
		this.keyCount ++
		this.valueCount ++

	}else{
		//key 已经存在
		if this.multyValuesSupported{
			whichMapValue.values = append(whichMapValue.values, value)
		}else{
			whichMapValue.values = []interface{}{value}
		}

		//更改计数
		this.valueCount ++
	}
}

func (this *MyMap) Get(key []byte) []interface{} {
	whichMapValue, _ := this.findKey(key)
	if nil == whichMapValue{
		return nil
	}else{
		return whichMapValue.values	//若为 nil 则说明当前 key 已经 remove 了
	}
}

//目前的实现并未将这个条目从 slot 中删除，只是清空了 key 所在的那个 mapvalue
func (this *MyMap) Remove(key []byte) []interface{}{
	whichMapValue, _ := this.findKey(key)
	if nil == whichMapValue{
		return nil
	}else{
		ret := whichMapValue.values

		//更改计数
		this.keyCount --
		this.valueCount -= len(whichMapValue.values)

		//执行 remove
		whichMapValue.values = nil	//标记当前值为 nil 即可. 后面的遍历, 或者 Get 时会知道这一点
		return ret
	}
}

//合并另外一个 map
func (this *MyMap) Merge(right *MyMap)  {
	for /*hashcode*/_, conflicts := range right.data{
		for _, conflict := range conflicts{
			key := conflict.key
			if nil == conflict.values{
				continue
			}
			for _, value := range conflict.values{
				//Put 中会更改 this 的计数
				this.Put(key, value)
			}
		}
	}
}

//遍历 MyMap, 同时对每一个 key-values 执行客户的回调. 若客户回调返回 false 则停止遍历
func (this *MyMap) Visit(visitor MyMapVisitor, vcount int, otherParams [] interface{}) int {

	count := 0
	for _,slot := range this.data{
		for _, mapValue := range slot{
			if count == vcount{
				break
			}

			//已删除当前键, 跳过它
			if nil == mapValue.values{
				continue
			}

			//遍历它
			count ++
			if !visitor.Visit(mapValue.key, mapValue.values, otherParams){
				break
			}
		}
	}
	return count
}

//有多少个不同的 key
func (this *MyMap) KeyCount() int {
	return this.keyCount
}

//有多少个不同的 value
func (this *MyMap) ValueCount() int {
	return this.valueCount
}

func (this *MyMap) Clear(){
	for _,slot := range this.data{
		for _, mapValue := range slot{
			mapValue.Clear()
		}
	}
	this.data = make(map[int] []myMapValue)

	this.keyCount = 0
	this.valueCount = 0
}

func (this *MyMap) Destroy(){
	for _,slot := range this.data{
		for _, mapValue := range slot{
			mapValue.Clear()
		}
	}
	this.data = nil

	this.keyCount = 0
	this.valueCount = 0
}

//是否有 hash 碰撞
func (this *MyMap) Stat()  {

	fmt.Println("keyCount: ", this.GetKeyCounts(), ", valueCount: ", this.GetValueCounts())

	stat := make(map[int]int)
	for _,slot := range this.data{
		stat[len(slot)] ++
	}
	conflict := false
	for k, v := range stat{
		if k > 1{
			fmt.Println(k, " : ", v)
			conflict = true
		}
	}
	if !conflict{
		fmt.Println("no hash conflict ~")
	}
}

/**
	hash = s[0]*31^(n-1) + s[1]*31^(n-2) + ... + s[n-1]
 */
func (this *MyMap) getHashCode(src []byte) int {
	h := 0
	if (len(src) > 0) {
		limit := len(src)
		for i:=0;i < limit;i ++{
			h = 31 * h + int(src[i])
		}
	}
	return h;
}

//------------------------------------------------------------------------
type MyMapVisitor interface {
	//若 Visit 返回 false 则停止遍历
	Visit(key []byte, values []interface{}, otherparams [] interface{}) bool
}

type DefaultMyMapVisitor struct {

}

func (this *DefaultMyMapVisitor) Visit(key []byte, values []interface{}, othreParams ... interface{}) bool {
	fmt.Println("key length: ", len(key), ", value counts: ", len(values))
	return true
}