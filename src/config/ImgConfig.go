package config

//含问题的大图
type ImgConfig struct {
	width int
	height int

	label1StartX int
	label1StartY int
	label1EndX int
	label1EndY int

	label2StartX int
	label2StartY int
	label2EndX int
	label2EndY int

	TheClipConfig *ClipConfig

	ClipIndexOffset int
	ClipIndexLength int

	Id uint8
}

var normalImgConfig = ImgConfig{
	width:293,
	height:190,

	label1StartX:120,
	label1StartY:0,
	label1EndX:176,
	label1EndY:29,

	label2StartX:180,
	label2StartY:0,
	label2EndX:236,
	label2EndY:29,
	TheClipConfig:&normalClipConfig,

	ClipIndexLength:4,
	ClipIndexOffset:normalClipConfig.SmallPicWidth * normalClipConfig.SmallPicHeight/2 - 8,

	Id:0,
}

func GetImgConfigById(id uint8) *ImgConfig {
	if id == 0{
		return &normalImgConfig
	}
	return nil
}

func GetImgConfigBySize(height, width int) *ImgConfig {
	if height==normalImgConfig.height && width == normalImgConfig.width {
		return &normalImgConfig
	}
	return nil
}