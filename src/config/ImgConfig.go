package config

//含问题的大图
type ImgConfig struct {
	width int
	height int



	TheClipConfig *ClipConfig

	OverrideClipLength int	//在计算 img index 时, 从每个 clip 提取 index 的长度是多少

	Id uint8
}

var normalImgConfig = ImgConfig{
	width:293,
	height:190,
	TheClipConfig:&normalClipConfig,

	OverrideClipLength: 1,//normalClipConfig.ClipLengh,

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