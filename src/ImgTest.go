package main

import (
	"bufio"
	"os"
	"fmt"
	"imgKmeans"
	"kmeans"
	"dbOptions"
	"imgOptions"
	"imgIndex"
)

func TestSaveClips()  {
	dbOptions.SaveAllClipsAsJpgOf([]byte("B996"),-1,-1)
}

func TestClipSmall()  {

	datas := ImgIndex.GetIndexOfClips("E:/clip/A332.jpg", 2000,10)



	fmt.Println("small pic count: ", len(datas), ", each index size: ", len(datas[0]))
	for _, line := range datas{
		fmt.Println(line)
	}
}

func ClipByPoint()  {
	left := "E:/clip/A332.jpg"
	x0 := 5
	y0 := 41
	x1 := 71
	y1 := 107
	ImgOptions.ImgClipByPoints(left,x0,y0,x1,y1,100)
}

func TestStatistics()  {
	left := "E:/merge/test/A332_small.jpg"
	right := "E:/merge/test/D777_small.jpg"
	stdin := bufio.NewReader(os.Stdin)
	threshold := 83
	for{
		fmt.Println("input statistic threshold: ")
		fmt.Fscan(stdin, &threshold)
		ImgOptions.Statistics(left, threshold)
		fmt.Println(left, " over `")
		ImgOptions.Statistics(right, threshold)
		fmt.Println(right, " over `")
	}

}

func TestMark()  {
	left := "E:/merge/test/A332_small.jpg"
	right := "E:/merge/test/D777_small.jpg"

	stdin := bufio.NewReader(os.Stdin)
	var threshold, k, factor , aroundSize int
	//var delta float64
	for {
		fmt.Print("please input threshold, k, factor, aroundSize: ")
		fmt.Fscan(stdin, &threshold, &k, &factor, &aroundSize)
		fmt.Println("threshold: ", threshold, " k: ", k,", factor: ", ", aroundSize: ", aroundSize)
		ImgOptions.ImageMarkCentersBlue(left,aroundSize, threshold, k ,factor)
		ImgOptions.ImageMarkCentersBlue(right,aroundSize, threshold, k ,factor)
	}

}

func Test5(){
	left := "E:/merge/test/A332_small.jpg"
	right := "E:/merge/test/D777_small.jpg"

	stdin := bufio.NewReader(os.Stdin)
	var threshold, k, factor int
	var delta float64
	for {
		fmt.Print("please input threshold, k, factor, delta: ")
		fmt.Fscan(stdin, &threshold, &k, &factor, &delta)
		fmt.Println("threshold: ", threshold, " k: ", k,", factor: ", ", delta: ", delta)
		ImgKmeans.DistanceOfImages(left, right , threshold , k, factor ,delta)

	}


}

func Test4()  {
	points, width, height := ImgKmeans.PickPoints("E:/merge/test/A332_small.jpg",120)
	stdin := bufio.NewReader(os.Stdin)
	var k, factor int
	for {
		fmt.Print("please input k: ")
		fmt.Fscan(stdin, &k, &factor)
		fmt.Println("k: ", k,", factor: ", factor, ", width: ", width, ", height: ",height)
		wndPoins := ImgKmeans.PickInitCenters(points,width,height,k,factor)
		fmt.Println(wndPoins)
	}
}

func Test3(){
	points := []kmeans.Point{
		{Entry:[]float64{1.0,3.0}}, {Entry:[]float64{1.0,4.0}}, {Entry:[]float64{1.0,5.0}},{Entry:[]float64{1.0,7.0}},
		{Entry:[]float64{2.0,3.0}}, {Entry:[]float64{2.0,4.0}},{Entry:[]float64{2.0,7.0}},
		{Entry:[]float64{3.0,5.0}},{Entry:[]float64{3.0,6.0}},{Entry:[]float64{3.0,7.0}},
		{Entry:[]float64{4.0,3.0}}, {Entry:[]float64{4.0,5.0}},{Entry:[]float64{4.0,6.0}},
	}



	stdin := bufio.NewReader(os.Stdin)
	var k, factor int
	for {
		fmt.Print("please input k: ")
		fmt.Fscan(stdin, &k, &factor)
		wndPoins := ImgKmeans.PickInitCenters(points,3,4,k,factor)
		fmt.Println(wndPoins)
	}

}

func Test2(){
	points := []kmeans.Point{
		{Entry:[]float64{1.0,3.0}}, {Entry:[]float64{1.0,4.0}}, {Entry:[]float64{1.0,5.0}},{Entry:[]float64{1.0,6.0}},{Entry:[]float64{1.0,7.0}},
		{Entry:[]float64{2.0,3.0}}, {Entry:[]float64{2.0,4.0}}, {Entry:[]float64{2.0,5.0}},{Entry:[]float64{2.0,6.0}},{Entry:[]float64{2.0,7.0}},
		{Entry:[]float64{3.0,3.0}}, {Entry:[]float64{3.0,4.0}}, {Entry:[]float64{3.0,5.0}},{Entry:[]float64{3.0,6.0}},{Entry:[]float64{3.0,7.0}},
		{Entry:[]float64{4.0,3.0}}, {Entry:[]float64{4.0,4.0}}, {Entry:[]float64{4.0,5.0}},{Entry:[]float64{4.0,6.0}},{Entry:[]float64{4.0,7.0}},
	}

	stdin := bufio.NewReader(os.Stdin)
	var leftx,lefty, rightx,righty float64
	for {
		fmt.Print("please input left x&y, right x&y :")
		fmt.Fscan(stdin, &leftx, &lefty, &rightx, &righty)

		fmt.Println(points)

		//leftTop := kmeans.Point{Entry:[]float64{leftx, lefty}}
		//rightBotom := kmeans.Point{Entry:[]float64{rightx, righty}}

		//ImgKmeans.PickInitCenters(points,3,4,3,)
	}

}

func Test() {
	//imgClipInDir("E:/clip/")
	//jpgDirCosSim("E:/clip/")
	//for i:=120;i <= 150;i+=1.0{
	//jpgBinaryzation("E:/clip/A332_small.jpg", i);
	//}
	//mergeJpgIndir("E:/merge/gongjiaoka/", "_merge_","_merge_", 100)
	//mergeEachDirInDir("E:/merge",[]string{"_binaray_","_merge_"},"_merge_",100)
	//	dirInDirBinarayzation("E:/merge/",[]string{"_binaray_","_merge_"},"_binaray_", 120)

	var k,factor,threshold int
	var delta float64
	stdin := bufio.NewReader(os.Stdin)
	for{
		fmt.Print("please input threshold, k, factor, delta: ")
		fmt.Fscan(stdin, &threshold, &k, &factor, &delta)

		fmt.Println("threshold:", threshold,", k:", k, ", factor: ", factor, " ,delta: ", delta)

		path := "E:/merge/test/"
		lefts := []string{"A332_small.jpg", "A364_small.jpg", "D777_small.jpg", "E257_small.jpg"}
		rights := []string{"D777_small.jpg", "A332_small.jpg", "A364_small.jpg", "D777_small.jpg"}

		if len(lefts) != len(rights){
			fmt.Println("make sure left file size is equal to the right")
			continue
		}

		for i, left := range lefts{
			dis := ImgKmeans.DistanceOfImages(path+left, path+rights[i], threshold, k,factor, delta)
			fmt.Println(left, " --- ", rights[i]," : ", dis)

		}
		fmt.Println("a case finished ~")
	}

	//mergeJpgs("E:/merge/test/",[]string{"A332_small_binaray_.jpg","C60_small_binaray_.jpg"},[]string{},"_merge_",100)
	//imgMatrix:=imgo.MustRead(dst)
	//imgMatrix_gray := imgo.Binaryzation(imgMatrix, 127)
	//gray_rgb := "E:/clip/gray_rgb.jpg"
	//err = imgo.SaveAsJPEG(gray_rgb, imgMatrix_gray, 50)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("%s generate finish\n", gray_rgb)


}
