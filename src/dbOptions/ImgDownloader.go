package dbOptions

import (
    "fmt"
    "net/url"
    //"strings"
    "net/http"
    "io/ioutil"
    "os"
    "strconv"
    "github.com/syndtr/goleveldb/leveldb"
    "runtime"
    "time"
    "math/rand"
    "config"
    "bufio"
    "strings"
    "errors"
)

var img_dir string = "E:/gen/3/"
var downloadFinished chan int


var invalidImgData []byte = []byte{255,216,255,225,0,24,69,120,105,102,0,0,73,73,42,0,8,0,0,0,0,0,0,0,0,0,0,0,255,236,0,17,68,117,99,107,121,0,1,0,4,0,0,0,30,0,0,255,238,0,14,65,100,111,98,101,0,100,192,0,0,0,1,255,219,0,132,0,16,11,11,11,12,11,16,12,12,16,23,15,13,15,23,27,20,16,16,20,27,31,23,23,23,23,23,31,30,23,26,26,26,26,23,30,30,35,37,39,37,35,30,47,47,51,51,47,47,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,1,17,15,15,17,19,17,21,18,18,21,20,17,20,17,20,26,20,22,22,20,26,38,26,26,28,26,26,38,48,35,30,30,30,30,35,48,43,46,39,39,39,46,43,53,53,48,48,53,53,64,64,63,64,64,64,64,64,64,64,64,64,64,64,64,255,192,0,17,8,0,190,1,37,3,1,34,0,2,17,1,3,17,1,255,196,0,106,0,1,0,3,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,2,3,4,7,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,16,1,0,2,2,2,1,3,2,2,10,3,0,0,0,0,0,0,1,2,17,3,33,18,4,49,65,19,81,34,20,5,97,113,129,145,177,50,66,98,178,35,161,114,21,17,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,255,218,0,12,3,1,0,2,17,3,17,0,63,0,244,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,28,51,231,236,136,139,116,140,76,223,214,113,197,63,122,127,29,121,215,55,138,86,49,53,136,158,217,143,187,235,196,99,14,75,205,102,177,19,29,102,107,121,172,77,162,35,54,182,39,214,61,225,166,218,79,77,241,170,184,138,252,83,17,235,30,158,208,13,231,205,219,91,117,152,165,184,137,205,45,51,28,218,43,244,87,241,247,248,117,223,20,139,108,237,252,211,49,31,108,254,167,61,126,104,237,73,156,234,172,230,211,49,21,140,214,241,206,113,244,38,45,248,93,52,142,209,107,68,241,31,247,227,216,29,81,231,205,166,34,177,174,177,136,156,218,248,142,103,219,136,77,188,219,215,124,211,166,117,247,141,125,179,207,105,114,215,239,156,204,78,98,218,227,158,103,139,78,125,161,107,71,111,46,98,38,102,209,190,147,215,251,98,57,144,116,87,203,219,91,90,155,181,244,183,91,94,188,251,71,213,156,126,99,121,215,54,138,214,38,58,226,38,102,115,152,206,35,17,42,205,43,243,204,106,191,120,174,187,197,175,121,204,86,211,159,86,63,28,244,219,120,188,77,41,120,152,154,255,0,84,204,91,211,244,243,0,236,167,155,121,213,222,107,89,183,104,172,215,54,174,51,30,249,170,149,252,195,100,205,230,105,76,83,251,249,227,215,219,149,52,210,127,15,49,61,166,109,106,197,235,72,205,163,164,98,98,127,94,25,210,179,109,155,107,254,200,204,253,211,90,115,207,49,233,60,96,29,246,242,107,74,199,120,180,90,99,180,226,182,152,143,219,133,60,111,50,187,169,89,180,76,94,211,137,197,103,175,175,215,209,79,43,101,239,174,158,62,188,252,155,99,238,137,226,98,177,28,254,172,169,227,222,52,108,174,34,105,163,119,17,91,127,69,227,219,246,130,218,252,237,187,54,90,148,213,158,177,51,235,143,73,199,172,175,226,249,118,223,49,91,83,172,245,237,152,158,39,156,56,254,253,86,190,123,82,243,89,138,68,113,51,54,188,226,27,248,58,190,45,151,174,201,159,155,92,99,174,99,172,214,121,140,112,14,246,123,118,90,147,72,172,68,246,182,39,51,142,62,177,245,52,109,249,181,70,206,179,94,222,210,195,206,196,91,199,180,251,109,175,63,160,20,159,204,54,198,222,159,23,24,206,62,236,255,0,131,77,158,93,169,186,53,68,83,159,234,181,241,237,158,99,217,193,52,183,197,55,181,98,38,211,61,62,218,243,207,183,191,252,54,223,242,124,241,108,78,107,21,199,31,72,137,144,111,167,206,157,187,99,94,43,25,152,143,230,137,246,159,79,220,173,63,48,189,235,218,43,174,185,246,182,200,137,254,12,60,109,119,137,241,230,98,121,156,199,57,226,34,115,198,56,83,197,188,71,90,108,217,241,83,164,219,56,140,231,183,233,137,7,124,121,87,141,86,217,107,106,251,102,35,237,180,204,115,245,152,137,101,255,0,161,110,243,76,83,136,207,108,219,31,225,150,27,123,219,70,250,210,103,102,190,212,138,91,17,25,247,159,72,231,148,236,174,221,91,39,102,206,209,22,174,34,45,178,51,56,254,48,14,200,242,47,104,164,214,43,126,209,51,49,22,231,143,76,118,194,150,243,55,124,148,215,93,81,155,231,31,117,125,185,159,73,150,26,226,41,62,20,236,142,179,141,145,57,143,105,142,50,174,155,106,175,195,120,196,99,117,162,113,244,183,160,59,227,116,207,147,58,122,241,20,237,219,246,225,179,24,157,127,137,152,207,251,58,70,99,19,233,153,247,108,0,0,0,0,0,0,0,0,0,8,48,144,17,136,147,9,1,24,18,2,48,142,177,244,88,4,96,194,64,70,4,128,140,68,250,152,143,95,116,128,129,32,32,194,64,70,4,128,131,9,1,2,64,66,64,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,31,255,217}

func initImgDB() *DBConfig {
    return GetImgDBWhichPicked()
}

func writeToDB(key string, value []byte)  {
    batch := leveldb.Batch{}
    batch.Put([]byte(key), value)

    imgDB := initImgDB()
    if nil == imgDB{
        return
    }
    imgDB.DBPtr.Write(&batch,&imgDB.WriteOptions)
}

func writeBatchToDB(batch *leveldb.Batch)  {
    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    imgDB.DBPtr.Write(batch,&imgDB.WriteOptions)
}

func readDBToImage(key string)  {

    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    value,err := imgDB.DBPtr.Get([]byte(key),&imgDB.ReadOptions)
    if err == leveldb.ErrNotFound{
        return
    }
    writeToFile(value, key)
}


func writeToAbsFile(content []byte, fileName string)  {
    image, err := os.Create(fileName)
    if err != nil {
        fmt.Println("create file failed:", fileName, err)
        return
    }

    defer image.Close()
    image.Write(content)
}

func isValidImage(img []byte) bool {
    if len(img)!=len(invalidImgData){
        return true
    }

    for i,byte := range invalidImgData{
        if byte!=img[i] {
            return true
        }
    }

    return false
}

func downLoadImageOnConnection(httpClient *http.Client, img_url string) (retData []byte, err error) {
    var response *http.Response
    response, err = httpClient.Get(img_url)

    if err != nil{
        return
    }

    if nil == response || 200 != response.StatusCode{
        err = errors.New("response code is not 200 or response is null")
        return
    }

    contenType := response.Header.Get("Content-Type")
    //text/html 或 image/jpeg 或其它
    if 0 != strings.Index(contenType, "image") {
        err = errors.New("not image")
        return
    }

    defer response.Body.Close()

    retData, err = ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("read data failed:", img_url, err)
        return
    }
    return
}

/***
    下载图片，在同一个 http 上下载多次图片会触发“请求太快，请稍后再试”，此图片对应的是 invalidImageData
 */
func saveImages(img_url string, goId int, maxImgId int, imgId int) int {


    proxy := func(_ *http.Request) (*url.URL, error) {
        //注意下面的 url 一定要包含 http:// 否则运行报错
        return url.Parse("http://proxy.webank.com:8080")
    }
    transport := &http.Transport{Proxy: proxy}
    client := &http.Client{Transport: transport, Timeout: time.Duration(20 * time.Second)}

    //r := rand.New(rand.NewSource(time.Now().UnixNano()))
    firstTime := true

    allowFailedCount := 3
    successDealed := 0
    dealCost := int64(0)

    for{
        startT := time.Now().Unix()

        if allowFailedCount <=0 {
            fmt.Println("current connection failed too many times, abort the connection")
            return imgId
        }

        data ,err := downLoadImageOnConnection(client, img_url)
        if err!=nil{
            allowFailedCount --
            continue
        }

        img_name := config.ThreadIdToName[goId] + strconv.Itoa(imgId)
        if !firstTime {
            if isValidImage(data){
                writeToDB(img_name,data)
                imgId ++
            }else{
                //如果触发了 invalidImage 则当前连接需要终止使用
                return imgId
            }
        }else{
            writeToDB(img_name,data)
            imgId ++
        }
        successDealed ++

    //    writeToFile(data, img_dir+"/"+img_name+".jpg")

        firstTime = false
        if(maxImgId == imgId){
            return imgId
        }

        endT := time.Now().Unix()
        dealCost += endT-startT

        if 0 != successDealed && successDealed % 100 == 0{
            fmt.Println("dealing ", successDealed, ", speed: 1/", float64(dealCost)/float64(successDealed),"s")
        }

    }

    return imgId
}


/***
    下载图片，在同一个 http 上下载多次图片会触发“请求太快，请稍后再试”，此图片对应的是 invalidImageData
 */
func singleSaveImages(img_url string, goId int, maxImgId int, imgId int) int {

    proxy := func(_ *http.Request) (*url.URL, error) {
        //注意下面的 url 一定要包含 http:// 否则运行报错
        return url.Parse("http://proxy.webank.com:8080")
    }
    transport := &http.Transport{Proxy: proxy}
    client := &http.Client{Transport: transport, Timeout: time.Duration(20 * time.Second)}

    data ,err := downLoadImageOnConnection(client,img_url)

    if err != nil {
        fmt.Println("download image error: ", err)
        return imgId
    }

    img_name := config.ThreadIdToName[goId] + strconv.Itoa(imgId)
    writeToDB(img_name,data)

  //  writeToFile(data, img_dir+"/"+img_name+".jpg")
    imgId ++

    return imgId
}



/**
    正常情况下，一次调用 saveImages 只会下载一张图片，但有时可以下载多张，我们利用这个特性加快下载
 */
func save(goId int, base int, times int)  {
    maxImageId := base+times
    nextImgId := base

    dealCost := int64(0)
    for i := 0; i != times; {

        if(0 != i && 0 == i % 100) {
            fmt.Println("dealing ", i, ", speed: 1/", float64(dealCost)/float64(i),"s")
        }
        startT := time.Now().Unix()
        newImgId := //saveImages("https://kyfw.12306.cn/passport/captcha/captcha-image?login_site=E&module=login", goId, maxImageId, nextImgId)
        singleSaveImages("https://kyfw.12306.cn/passport/captcha/captcha-image?login_site=E&module=login", goId, maxImageId, nextImgId)
        endT := time.Now().Unix()

        dealCost += endT - startT

        i += (newImgId - nextImgId)
        nextImgId = newImgId

        if(nextImgId >= maxImageId){
            fmt.Println("just before thread ", goId, " going to return")
            //提前完成任务: 此时  nextImgId 正好是它下载图片的个数
            downloadFinished <- nextImgId-base
            return
        }
    }
    fmt.Println("just before thread ", goId, " going to return")
    //此时  nextImgId 正好是它下载图片的个数
    downloadFinished <- nextImgId
}

func HowManyImages() int  {

    imgDB := initImgDB()
    if nil == imgDB{
        return 0
    }

    iter := imgDB.DBPtr.NewIterator(nil, &imgDB.ReadOptions)

    if(!iter.First()){
        fmt.Println("seek to first error")
    }
    count := 0
    for iter.Valid(){
        //writeToFile(iter.Value(), string(iter.Key()))
        curKey := iter.Key()
        if 0!=len(curKey){
            b := curKey[0]
            if b>= 'A' && b <= 'Z'{
                count ++
            }
        }
        iter.Next()
    }
    fmt.Println(count," images in total")
    return count
}

func HowManyImageClips() int  {

    imgDB := InitImgClipsDB()
    if nil == imgDB{
        return 0
    }

    iter := imgDB.DBPtr.NewIterator(nil, &imgDB.ReadOptions)

    if(!iter.First()){
        fmt.Println("seek to first error")
    }
    count := 0
    for iter.Valid(){
        //writeToFile(iter.Value(), string(iter.Key()))
        curKey := iter.Key()
        if 0!=len(curKey){
            b := curKey[0]
            if b>= 'A' && b <= 'Z'{
                count ++
            }
        }
        iter.Next()
    }
    fmt.Println(count," clips in total")
    return count
}


func Begin(lastBase int, cores int, eachTimes int) int {
    runtime.GOMAXPROCS(cores)
    downloadFinished = make(chan int, cores)

    for i:=0;i < cores; i++{
        go save(i,lastBase, eachTimes)
        fmt.Println("thread ", i, " going to start")
    }

    //等待线程全部执行完毕
    total := 0
    for i:=0;i < cores;i ++{
        f := <- downloadFinished
        total += f
        fmt.Println("thread ", i ," finished")
    }

    fmt.Println("total download: ", total)
    return total
}


func setStatInfo(lastBase int, cores int, eachTimes int, costSecs int64)  {

    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    db := imgDB.DBPtr

    writeOptions := &imgDB.WriteOptions

    db.Put([]byte("base"), []byte(strconv.Itoa(lastBase)), writeOptions)

    db.Put([]byte("cores"), []byte(strconv.Itoa(cores)), writeOptions)

    db.Put([]byte("eachTimes"), []byte(strconv.Itoa(eachTimes)), writeOptions)

    db.Put([]byte("costSecs"), []byte(strconv.FormatInt(costSecs, 10)), writeOptions)

    key := "stat." + time.Now().Format("2006-01-02 15:04:05");
    value := "base: " + strconv.Itoa(lastBase) + ", cores:" + strconv.Itoa(cores) + ", eachTimes:" + strconv.Itoa(eachTimes) + ", cost:"+strconv.FormatInt(costSecs, 10)+"s"
    db.Put([]byte(key),[]byte(value), writeOptions)
    db.Put([]byte("current_stat"), []byte(key),  writeOptions)
}

func GetStatInfo() (lastBase int, lastCoers int, lastEachTimes int, lastCostScs int64, remark string){

    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    db := imgDB.DBPtr
    readOptions := &imgDB.ReadOptions

    {
        cs,err := db.Get([]byte("current_stat"), readOptions)
        if err==leveldb.ErrNotFound{
            remark = "no stat data"
        }else{
            ret, err:= db.Get([]byte(cs), readOptions)
            if err == leveldb.ErrNotFound{
                remark = "no stat data"
            }else{
                remark = string(ret)
            }
        }
    }


    {
        lb, err := db.Get([]byte("base"), readOptions)
        if err == leveldb.ErrNotFound{
            lastBase = 0
        }else{
            ilb,err:=strconv.Atoi(string(lb))
            if err==nil{
                lastBase = ilb
            }else{
                lastBase =0
            }
        }
    }


    {
        lc, err := db.Get([]byte("cores"), readOptions)
        if err == leveldb.ErrNotFound{
            lastCoers = 0
        }else{
            ilc,err:=strconv.Atoi(string(lc))
            if err==nil{
                lastCoers = ilc
            }else{
                lastCoers =0
            }
        }
    }

    {
        let, err := db.Get([]byte("eachTimes"), readOptions)
        if err == leveldb.ErrNotFound{
            lastEachTimes = 0
        }else{
            ilet,err:=strconv.Atoi(string(let))
            if err==nil{
                lastEachTimes = ilet
            }else{
                lastEachTimes =0
            }
        }
    }

    {
        lcost, err := db.Get([]byte("costSecs"), readOptions)
        if err == leveldb.ErrNotFound{
            lastCostScs = 0
        }else{
            icost,err:=strconv.ParseInt(string(lcost),10,64)
            if err==nil{
                lastCostScs = icost
            }else{
                lastCostScs =0
            }
        }
    }
    return
}

func ImgDBStatRepair()  {

    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    db := imgDB.DBPtr
    writeOptions := &imgDB.WriteOptions

    db.Delete([]byte("lastBase"), writeOptions)
    db.Delete([]byte("lastTime"), writeOptions)


    keySize := HowManyImages()
    db.Put([]byte("base"),[]byte(strconv.Itoa(keySize)), writeOptions)
    db.Put([]byte("cores"),[]byte("8"), writeOptions)
    db.Put([]byte("eachTimes"),[]byte("1000"), writeOptions)
}

func RandomVerify(){

    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    db := imgDB.DBPtr
    readOptions := &imgDB.ReadOptions


    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i:=0; i< 40; i++ {
        index := r.Intn(1000)

        letter:=r.Intn(8)

        key := config.ThreadIdToName[letter] + strconv.Itoa(index);
        value,err := db.Get([]byte(key), readOptions)
        if err == leveldb.ErrNotFound{
            fmt.Println("can't get :", key)
        }else {
            writeToFile(value, "E:/gen/verify/" + key+".jpg")
        }
    }
}

func Dump(dir string, count int)  {
    imgDB := initImgDB()
    if nil == imgDB{
        return
    }

    db := imgDB.DBPtr
    readOptions := &imgDB.ReadOptions

    iter := db.NewIterator(nil, readOptions)

    if(!iter.First()){
        fmt.Println("seek to first error")
    }

    for iter.Valid(){
        if count <= 0{
            break
        }
        writeToAbsFile(iter.Value(), dir + "/" + string(iter.Key()) + ".jpg")
        iter.Next()
        count --
    }
}

func DownloaderRun()  {
    cores := 8
    eachTimes := 10000
    db := 8

    stdin := bufio.NewReader(os.Stdin)
    for{
        fmt.Print("select a img db to store image: ")
        fmt.Fscan(stdin, &db)
        fmt.Print("input thread num and each iter num: ")
        fmt.Fscan(stdin, &cores, &eachTimes)
        fmt.Println("img db: ", db,", thread num: ", cores, ", iter num: ", eachTimes)

        imgDB := PickImgDB(db)

        fmt.Println("cores: ", cores, ", eachTimes: ", eachTimes)

        newBase, _ ,_, _, _:= GetStatInfo()

        st := time.Now().Unix()

        total := Begin(newBase, cores, eachTimes)
        et := time.Now().Unix()
        fmt.Println("cost ", (et-st), " seconds")

        setStatInfo(newBase + total,cores,eachTimes,(et-st))

        imgDB.CloseDB()
    }




}