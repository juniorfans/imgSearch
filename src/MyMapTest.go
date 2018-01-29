package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
	"time"
	"math/rand"
	"util"
	"imgCache"
)

func FunctionTest() {
	myMap := imgCache.NewMyMap(true)
	stdin := bufio.NewReader(os.Stdin)
	var opt string
	var keystr, value string
	for {
		fmt.Print("input add/remove key value: ")
		fmt.Fscan(stdin, &opt, &keystr, &value)
		keyByteStrs := strings.Split(keystr, ",")
		keyBytes := make([]byte, len(keyByteStrs))
		for i, bs := range keyByteStrs {
			ibs, _ := strconv.Atoi(bs)
			keyBytes[i] = byte(ibs)
		}
		if 0 == strings.Compare("add", opt) {
			fmt.Println("add ", keyByteStrs, "-", value)
			myMap.Put(keyBytes, value)
		} else if 0 == strings.Compare("remove", opt) {
			fmt.Println("remove ", keyByteStrs)
			myMap.Remove(keyBytes)
		} else {
			fmt.Println("invalid option: ", opt)
			continue
		}

		fmt.Println("get ", keyByteStrs, "-", myMap.Get(keyBytes))
	}
}

func randomBytes(length int) []byte {
	ret := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i ++ {
		ret[i] = byte(r.Intn(256))
	}
	return ret
}

func randomKeyLen() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(1000)
}

func randomValueLen() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(1000)
}

func incrementBytes(src []byte) []byte {
	ret := fileUtil.CopyBytesTo(src)
	for i,_ := range ret{
		if ret[i] < 255{
			ret[i] += 1
			break
		}
	}
	return ret
}

func BenchTestMyMap() {
	myMap := imgCache.NewMyMap(true)
	testCase := 1000
	btime := time.Now().UnixNano()

	key := randomBytes(100)
	//baseKey := fileUtil.CopyBytesTo(key)
	//incrementCount := 0
	value := randomBytes(randomValueLen())
	for i:=0 ; i < testCase; i ++{
	/*	if incrementCount >= 0{
			key = incrementBytes(key)
			incrementCount = 0
		}else{
			incrementCount ++
			//stay key same
		}
*/
		key = incrementBytes(key)
		value = incrementBytes(value)
		myMap.Put(key, value)
	//	myMap.Remove(key)
	}
	etime := time.Now().UnixNano()

	/*
	key = fileUtil.CopyBytesTo(baseKey)
	for i:= 1000;i < 2000;i ++{
		key = incrementBytes(key)
		values := myMap.Get(key)
		if 0 < len(values){
			fmt.Println("found ", len(values), ", first value len: ", len(values[0].([]byte)))

		}else{
			fmt.Println("error, not found")
		}
	}
*/
	fmt.Println("test cost : ", (etime - btime)/1000000, " ms")

	myMap.Stat()
}

func BenchTestMap() {
	mapRes := make(map[int]int)
	testCase := 10000000
	btime := time.Now().Unix()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for ; testCase != 0; testCase -- {
		key := r.Int()
		mapRes[key] = r.Int()
		delete(mapRes, key)
		if testCase & 10000 == 0 {
			//	fmt.Println("map count: ", len(mapRes))
		}
	}
	etime := time.Now().Unix()
	fmt.Println("test cost : ", (etime - btime))
}

func getHashCode(src []byte) int {
	h := 0
	if (len(src) > 0) {
		for _, c := range src {
			h = 31 * h + int(c)
		}
	}
	return h;
}

func getHashCodeEx(src []byte) int {
	h := 0
	if (len(src) > 0) {
		for _, c := range src {
			//h = 31 * h + int(c)
			h = (h << 5)-h + int(c)
		}
	}
	return h;
}

func TestHashBench()  {
	testCase := 1000000
	st := time.Now().UnixNano()
	for ;testCase != 0;testCase --{
		getHashCode(randomBytes(randomKeyLen()))
	}
	et := time.Now().UnixNano()
	fmt.Println("test cost ", (et - st)/1000000, " ms")
}

func main() {
	BenchTestMyMap()
}
