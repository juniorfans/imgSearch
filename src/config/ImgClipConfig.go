package config

type ClipConfig struct {
	SmallPicWidth int
	SmallPicHeight int
	StartOffsetX int
	StartOffsetY int
	IntervalXBetweenSmallPic int
	IntervalYBetweenSmallPic int
	bigPicWidth int
	bigPicHeight int
	SmallPicCountInX int
	SmallPicCountInY int
	Id uint8
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
	bigPicWidth:293,
	bigPicHeight:190,
	SmallPicCountInX: 4,
	SmallPicCountInY: 2,
	Id : 0,
}

func GetClipConfigById(id uint8) *ClipConfig {
	if id == 0{
		return &normalClipConfig
	}
	return nil
}

func GetClipConfigBySize(height, width int) *ClipConfig {
	if height==normalClipConfig.bigPicHeight && width == normalClipConfig.bigPicWidth {
		return &normalClipConfig
	}
	return nil
}