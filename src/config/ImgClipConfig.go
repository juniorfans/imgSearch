package config


//注意当子图数量发生变化时此处要变
var CLIP_COUNTS_OF_IMG = uint8(8)

type ClipConfig struct {
	SmallPicWidth            int
	SmallPicHeight           int
	StartOffsetX             int
	StartOffsetY             int
	IntervalXBetweenSmallPic int
	IntervalYBetweenSmallPic int
	BigPicWidth              int
	BigPicHeight             int
	SmallPicCountInX         int
	SmallPicCountInY         int
	Id                       uint8
	ImgConfigId              uint8

	ClipOffsets              []int
	ClipLengh                int

	//-----------------
	cachedClipLeftTopPoints  []Point
}

/**
	(5,41) 是第一个子图的左上角像素

 */
var normalClipConfig = ClipConfig{
	SmallPicWidth: 67,
	SmallPicHeight: 67,
	StartOffsetX: 5,
	StartOffsetY: 41,
	IntervalXBetweenSmallPic:5,
	IntervalYBetweenSmallPic:5,
	BigPicWidth:293,
	BigPicHeight:190,
	SmallPicCountInX: 4,
	SmallPicCountInY: 2,

	//回字型采样
	ClipOffsets: []int{	10*67+10, 11*67-10, 57*67+10, 58*67-10,
				20*67+20, 21*67-20, 47*67+20, 48*67-20,
				30*67+30, 31*67-30, 37*67+30, 38*67-30,
				},
	ClipLengh: 2,

	Id : 0,
}

type Point struct {
	x, y int
}

/**
	子图序列是
	0	2	4	6
	1	3	5	7
 */
func (this *ClipConfig) GetClipsLeftTop() []Point {
	if 0 != len(this.cachedClipLeftTopPoints){
		return this.cachedClipLeftTopPoints
	}
	ret := make([]Point, 8)
	xStep := this.SmallPicWidth+this.IntervalXBetweenSmallPic
	yStep := this.SmallPicHeight+this.IntervalYBetweenSmallPic

	xLimit := this.BigPicWidth -this.IntervalXBetweenSmallPic-this.SmallPicWidth
	yLimit := this.BigPicHeight -this.IntervalYBetweenSmallPic-this.SmallPicHeight
	ci := 0
	for i:=this.StartOffsetX;i<= xLimit; i+=xStep {
		for j := this.StartOffsetY; j <= yLimit; j += yStep {
			//i,j 为左上角
			ret[ci] = Point{x:i,y:j}
			ci ++
		}
	}
	this.cachedClipLeftTopPoints = ret
	return ret
}

/**
	大图中的 x,y 位于哪个子图里面
 */
func (this *ClipConfig) WhichClip (x,y int) uint8 {
	points := this.cachedClipLeftTopPoints
//	y += this.StartOffsetY
	for i, point := range points{
		if x >= point.x && y >= point.y && x <= point.x+this.SmallPicWidth && y <= point.y+this.SmallPicHeight{
			return uint8(i)
		}
	}
	return 255
}

func GetClipConfigById(id uint8) *ClipConfig {
	if id == 0{
		return &normalClipConfig
	}
	return nil
}

func GetClipConfigBySize(height, width int) *ClipConfig {
	if height==normalClipConfig.BigPicHeight && width == normalClipConfig.BigPicWidth {
		return &normalClipConfig
	}
	return nil
}