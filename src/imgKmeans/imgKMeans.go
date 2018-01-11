package ImgKmeans

import (
	//	"github.com/salkj/kmeans"
	"kmeans"
	"sort"
	"github.com/Comdex/imgo"
	"fmt"
)

type CPoints []kmeans.Point

func (a CPoints) Len() int {
	return len(a)
}
func (a CPoints) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
//先比较横坐标，再比较纵坐标
func (a CPoints) Less(i, j int) bool {

	if a[i].Entry[0] == a[j].Entry[0] {
		return a[i].Entry[1] < a[j].Entry[1]
	} else {
		return a[i].Entry[0] < a[j].Entry[0]
	}
}

func DistanceOfImages(left, right string, threshold int, k, factor int, delta float64) float64 {
	//(fileName string,threshold, k ,factor int, delta float64)
	leftCenter := ImageKmeans(left, threshold, k,factor, delta)
	rightCenter := ImageKmeans(right, threshold, k, factor,delta)
	return DistanceCal(leftCenter, rightCenter)
}

func DistanceCal(lefts, rights []kmeans.Point) float64 {
	if 0 == len(lefts) && 0 == len(rights) {
		return float64(0)
	}

	sort.Sort(CPoints(lefts))
	sort.Sort(CPoints(rights))

	sum := float64(0)
	for e:=0;e < len(lefts);e++{
		sum += (lefts[e]).DistanceTo(rights[e])
	}
	return sum/float64(len(lefts))
}



func Distance(lefts, rights []kmeans.Point) float64 {
	if 0 == len(lefts) && 0 == len(rights) {
		return float64(0)
	}

	sort.Sort(CPoints(lefts))
	sort.Sort(CPoints(rights))

	llen := len(lefts)
	rlen := len(rights)
	minLen := llen
	if (llen > rlen) {
		minLen = rlen
	}
	i := 0;
	for {
		if (i < minLen) {
			fmt.Println("left: ", lefts[i], " ----- right: ", rights[i])
		} else {
			if (i < llen) {
				fmt.Println("left: ", lefts[i])
			} else if (i < rlen) {
				fmt.Println("right: ", rights[i])
			} else {
				break
			}

		}
		i++
	}

	return float64(0)
}

type wndPoint struct {
	leftTop    kmeans.Point
	Center     kmeans.Point
	pointCount int
}
type CwndPoints []wndPoint

func (a CwndPoints) Len() int {
	return len(a)
}
func (a CwndPoints) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
//先比较横坐标，再比较纵坐标
func (a CwndPoints) Less(i, j int) bool {

	return a[i].pointCount > a[j].pointCount
}

/**

 */
func PickInitCenters(points []kmeans.Point, width, height ,k ,wndFactor int) CwndPoints {
	sort.Sort(CPoints(points))

	wps := CwndPoints{}

	wndX := width / wndFactor
	wndY := height / wndFactor
	if 0 == wndX{
		fmt.Println("wndFactor is too big, set windows.x size=1")
		wndX = 1
	}
	if 0 == wndY{
		fmt.Println("wndFactor is too big, set windows.y size=1")
		wndY = 1
	}

//	fmt.Println("wndx: ", wndX, "wndy: ", wndY)

	y := 0
	for ;y < height ;y += wndY{
		x := 0
		for ;x < width;x+= wndX{
			//(x,y) 为窗口的左上角, (x+wndX, y+wndY)为窗口的右下角
			leftTop := kmeans.Point{Entry:[]float64{float64(x), float64(y)}}
			rightBotom := kmeans.Point{Entry:[]float64{float64(x+wndX), float64(y+wndY)}}
			curCenter := kmeans.Point{Entry: []float64{float64(x)+ float64(wndX)/2, float64(y)+float64(wndY)/2}}
			count := CountPointsInWnd(points, leftTop, rightBotom)
		//	fmt.Println("query in ", leftTop, " --- ", rightBotom, ", count: ", count)

			wp := wndPoint{leftTop: leftTop, pointCount:count, Center:curCenter}
			wps = append(wps, wp)
		}
	}

	sort.Sort(wps)

	//fmt.Println(wps)
	if len(wps) > k{
		return wps[0:k]
	}else{
		return wps
	}
}



/*
	找出 points 中落在 (leftTop, rightBotom) 窗口内的点有哪些
 */
func CountPointsInWnd(points []kmeans.Point, leftTop, rightBotom kmeans.Point) (retCounts int ){
	inCount := 0

	for _,point := range points{
		if pointBigThan(point, leftTop) && pointLessThan(point, rightBotom){
			//fmt.Println(point)
			inCount ++
		}
	}
	retCounts = inCount

	return
}


func pointLessThan(left, right kmeans.Point) bool {
	return left.Entry[1] <= right.Entry[1] && left.Entry[0] <= right.Entry[0]
}

func pointEqual(left, right kmeans.Point) bool {
	return left.Entry[0] == right.Entry[0] && left.Entry[1] == right.Entry[1]
}

func pointBigThan(left, right kmeans.Point) bool {
	return left.Entry[1] >= right.Entry[1] && left.Entry[0] >= right.Entry[0]
}

/**
	先将 fileName 二值化，选出二值化后有值的那些点
 */
func PickPoints(fileName string, threshold int) (retPoints []kmeans.Point, width, heigth int) {
	imgMatrixLeft, err := imgo.Read(fileName)

	if nil != err {
		fmt.Println("open file error: ", fileName, ": ", err)
		return
	}

	imgMatrixLeft = imgo.RGB2Gray(imgMatrixLeft)


	heigth = len(imgMatrixLeft)        //高 - 即长方形中的宽
	width = len(imgMatrixLeft[0])      //宽 - 即长方形中的长

	data := []kmeans.Point{}

	for i := 0; i < heigth; i++ {
		for j := 0; j < width; j++ {
			r := int(imgMatrixLeft[i][j][0]) + int(imgMatrixLeft[i][j][1]) + int(imgMatrixLeft[i][j][2]) / 3
			if r <= threshold {
				//	fmt.Println(i," ",j," ","(",imgMatrixLeft[i][j][0],",",imgMatrixLeft[i][j][1],",",imgMatrixLeft[i][j][2],")")
				//将 (j, i) 作为点加入进去, 横坐标，纵坐标
				data = append(data, kmeans.Point{[]float64{float64(j), float64(i)}})
			}

			imgMatrixLeft[i][j][0] = 255
			imgMatrixLeft[i][j][1] = 255
			imgMatrixLeft[i][j][2] = 255
		}
	}
	retPoints = data
	return ;
}



func ImageKmeans(fileName string,threshold, k ,factor int, delta float64) []kmeans.Point {

	//	fmt.Println("height: ", height, ", width: ", width)

	data,width,height := PickPoints(fileName, threshold)

	fmt.Println("kmeas poins: ", len(data))

	//	fmt.Println("point size: ", len(data))

	wndPoins := PickInitCenters(data,width,height,k,factor)


	cps := make(CPoints, len(wndPoins))
	initCenters := make([]kmeans.Centroid, len(wndPoins))
	for i, wp := range wndPoins{
		initCenters[i].Center = wp.Center
		cps[i].Entry = wp.Center.Entry
	}

	sort.Sort(cps)
//	fmt.Println("init center: ", cps)

	//fmt.Println(wndPoins)
	//fmt.Println(initCenters)
	var centers []kmeans.Centroid = kmeans.KMEANS(data, initCenters, k, float64(delta))

	ret := make([]kmeans.Point, len(centers))
	for i, c := range centers{
		//fmt.Println("center: ", c.Center)
		ret[i]=c.Center
	}
	return  ret

	/*
	color := uint8(0)
	eachAdd := uint8(255 / len(centers))
	for i, c := range centers{
		c.Center
		points := c.Points
		if len(points) == 0{
			continue
		}

		for _,point := range points{
			if len(point.Entry) == 0{
				continue
			}
			x := uint8(point.Entry[1])
			y := uint8(point.Entry[0])
			fmt.Println("x:",x,",y:",y)

			imgMatrixLeft[x][y][i % 3]=color

			imgMatrixLeft[x][y][3]=0
		//	imgMatrixLeft[x][y][1]=color
		//	imgMatrixLeft[x][y][2]=color
		}

		color += eachAdd
	}

	newName := strings.Replace(fileName, ".", suffixToAdd + "_" + strconv.Itoa(k) + "_" + strconv.FormatFloat(delta,'f',5,64) + ".", 1)
	err = imgo.SaveAsJPEG(newName, imgMatrixLeft, 100)
	if nil == err{
		fmt.Println("kmenas success, ", newName)
	}else{
		fmt.Println("kmeans failed, save jpg error. ", err)
	}
	*/
}
