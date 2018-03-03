package imgCache

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"sync"
)

/**
	提供一个以 []byte 为键, []interface{} 为值的 map, 非线程安全
 */


/**
	MyConcurrentMapValue 作为 MyConcurrentMap 的元素, 它保持了使用者传入的 key-value, 用于准确定位
 */
type MyConcurrentMapValue struct {
	values []interface{}
	key []byte
}

func (this *MyConcurrentMapValue) Clear()  {
	this.key = nil
	this.values = nil
}

type MyConcurrentMap struct {

	state	[1000000]int32	//若 state[i] 为0, 则表示所有模除 1000000 等于 i 的 hash, 均可以读写. 否则需要等待
					//当 hash 碰撞时, 采用拉链法解决
	data sync.Map//map[int] []MyConcurrentMapValue

	multyValuesSupported bool

					//-------------------
	keyCount, valueCount int	//不同的 key, 不同的 value 的个数
}

func NewMyConcurrentMap(multyValuesSupported bool) *MyConcurrentMap {
	ret := MyConcurrentMap{}
	ret.multyValuesSupported = multyValuesSupported
	ret.keyCount = 0
	ret.valueCount = 0
	return &ret
}

func (this *MyConcurrentMap) KeySet() [][]byte {
	var ret [][]byte

	this.data.Range(func(key, interfaceValue interface{}) bool {
		values := interfaceValue.([]MyConcurrentMapValue)
		for _,value := range values{
			ret = append(ret, value.key)
		}
		return true
	})

	return ret
}


func (this *MyConcurrentMap) findKey(hash int, key []byte) (ret *MyConcurrentMapValue){

	ret = nil
	if  hashConflicts, ok :=this.data.Load(hash); ok{
		//当前 byte 的 hash 已存在槽, hash 存在碰撞, 碰撞意味着 hash 一样，但 key 可能一样/不一样的元素.
		//下面依次遍历该槽中所有的 MyConcurrentMapValue, 当找到了相同的 key 时, 则合并/替代
		values := hashConflicts.([]MyConcurrentMapValue)
		if 1<len(values){
			//		fmt.Println("occ more :", len(hashConflicts))
		}

		for i,conflict := range values {
			if bytes.Equal(conflict.key, key){
				ret = &values[i]
				break
			}
		}
	}
	return
}

//从 MyConcurrentMap 中删除一个键, 注意不能直接删除 hash 所在的那个槽, 因为不同的 key 可能算得同样的 hash
func (this *MyConcurrentMap) removeKey(key []byte, hash int)  {
	hashConflicts, ok :=this.data.Load(hash)
	if !ok{
		return
	}
	values := hashConflicts.([]MyConcurrentMapValue)
	if 0 == len(values){
		this.data.Delete(hash)
		return
	}

	fnd := -1
	for i,value := range values{
		if bytes.Equal(value.key, key){
			fnd = i
			break
		}
	}

	if -1 != fnd{
		newValues := make([]MyConcurrentMapValue, len(values) - 1)
		ci := 0
		if fnd != 0{
			ci += copy(newValues[ci:], values[:fnd])
		}

		if fnd != len(values)-1{
			ci += copy(newValues[ci:], values[fnd+1:])
		}
		this.data.Store(hash, newValues)
	}
}

func (this *MyConcurrentMap) waitRWFor(hash int)  {
	//自旋锁 0: 可写入, 可读写的状态; 1: 拒绝读写, 将自旋等待
	if hash<0{
		hash = -hash
	}

	for !atomic.CompareAndSwapInt32(&(this.state[hash % 1000000]), 0, 1){
		fmt.Println("waiting, hash: ", hash)
	}
}

func (this *MyConcurrentMap) endWaitRWFor(hash int)  {
	//自旋锁 0: 可写入, 可读写的状态; 1: 拒绝读写
	if hash<0{
		hash = -hash
	}
	for !atomic.CompareAndSwapInt32(&(this.state[hash % 1000000]), 1, 0){
		fmt.Println("end wait error, other thread enter in, hash: ", hash)
	}
}

/**
	put key-value into the map. will not copy memory, only hold reference. so if change the slice outside,
	everything goes wrong: hashcode won't calc again
 */
func (this *MyConcurrentMap)Put(key []byte, value interface{})  {
	hash := this.getHashCode(key)
	this.waitRWFor(hash)
	defer this.endWaitRWFor(hash)

	whichMapValue := this.findKey(hash, key)

	if nil == whichMapValue{
		//key 不存在
		mapValue := MyConcurrentMapValue{key:key, values:[]interface{}{value}}
		interfaceValues, ok := this.data.Load(hash)

		if !ok{
			this.data.Store(hash, []MyConcurrentMapValue{mapValue})
		}else{
			values := interfaceValues.([]MyConcurrentMapValue)
			this.data.Store(hash, append(values, mapValue))
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

func (this *MyConcurrentMap) Contains(key []byte) bool {
	hash := this.getHashCode(key)
	this.waitRWFor(hash)
	defer this.endWaitRWFor(hash)

	whichMapValue := this.findKey(hash, key)

	return nil != whichMapValue
}


func (this *MyConcurrentMap) Get(key []byte) []interface{} {
	hash := this.getHashCode(key)
	this.waitRWFor(hash)
	defer this.endWaitRWFor(hash)

	whichMapValue := this.findKey(hash, key)

	if nil == whichMapValue{
		return nil
	}else{
		return whichMapValue.values	//若为 nil 则说明当前 key 已经 remove 了, 或者不存在
	}
}

//目前的实现并未将这个条目从 slot 中删除，只是清空了 key 所在的那个 mapvalue //这一点已经优化掉了, 真正的删除了键
//
func (this *MyConcurrentMap) Remove(key []byte) []interface{}{
	hash := this.getHashCode(key)

	this.waitRWFor(hash)
	defer this.endWaitRWFor(hash)

	whichMapValue := this.findKey(hash, key)

	if nil == whichMapValue{
		return nil
	}else{
		ret := whichMapValue.values

		//更改计数
		this.keyCount --
		this.valueCount -= len(whichMapValue.values)

		//执行 remove
		whichMapValue.values = nil	//标记当前值为 nil 即可. 后面的遍历, 或者 Get 时会知道这一点 //这点已经优化, 见下面

		//注意: 不能直接删除 hash 所在的那个槽, 因为不同的 key 可能会计算得到相同的 hash
		this.removeKey(key, hash)
		return ret
	}
}

//有多少个不同的 key
func (this *MyConcurrentMap) KeyCount() int {
	return this.keyCount
}

//有多少个不同的 value
func (this *MyConcurrentMap) ValueCount() int {
	return this.valueCount
}


//是否有 hash 碰撞
func (this *MyConcurrentMap) Stat()  {

	fmt.Println("keyCount: ", this.KeyCount(), ", valueCount: ", this.ValueCount())

	stat := make(map[int]int)

	this.data.Range(func(key, value interface{}) bool {
		stat[key.(int)] ++
		return true
	})

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
func (this *MyConcurrentMap) getHashCode(src []byte) int {
	h := 0
	if (len(src) > 0) {
		limit := len(src)
		for i:=0;i < limit;i ++{
			h = 31 * h + int(src[i])
		}
	}
	return h;
}
