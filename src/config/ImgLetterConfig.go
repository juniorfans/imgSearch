package config

type Region struct {
	LTX, LTY, RBX, RBY int
}

type LetterConfig struct {
	Id uint8
	ImgConfigId uint8	//所属的 img 配置的 id

	Width, Height int

	Regions []Region
	IndexLength int
	IndexOffset int
}

var normalLetterConfig = LetterConfig{
	Id : 0,
	ImgConfigId :0,
	Width :57,
	Height : 30,
	Regions:[]Region{Region{120,0,176,29}}, //, Region{180,0,236,29}},
	IndexLength: 60,	//-1 表示所有
	IndexOffset: (57/2)*30 + 15,
}

func GetLetterConfigById(id uint8) *LetterConfig {
	if id == 0{
		return &normalLetterConfig
	}
	return nil
}