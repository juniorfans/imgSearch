package main

import (
	"imgSearch/src/imgCache"
	"fmt"
	"imgSearch/src/util"
	"time"
	"math/rand"
)

var sig chan int

func main()  {
	//startTest()
	BenchTestMyConcurrentMap()
}

func startTest()  {
	sig = make(chan int, 10)
	cmap := imgCache.NewMyConcurrentMap(false)
	for i:=0;i < 10;i ++{
		go test(i, cmap)
	}

	for i:=0;i < 10;i ++{
		threadId := <- sig
		fmt.Println(threadId, " finished~")
	}

	cmap.Stat()

}

func test(threadId int, myMap *imgCache.MyConcurrentMap)  {
	key := []byte{0,0,0,0}
	for i:=0;i < 255 ;i ++{
		fileUtil.BytesIncrement(key)
		myMap.Put(key, i)
	}
	sig <- threadId
}



func BenchTestMyConcurrentMap() {
	myMap := imgCache.NewMyConcurrentMap(true)
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