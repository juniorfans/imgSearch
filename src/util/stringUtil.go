package fileUtil

func BytesEqual(left, right []byte) bool {
	if nil == left {
		return nil == right
	}else if nil == right{
		return false
	}else{
		if len(left) != len(right){
			return false
		}
		for i, cr := range right{
			if left[i] != cr{
				return false
			}
		}
		return true
	}
}

//left starts with right
func BytesStartWith(left, right []byte) bool {
	if nil == left {
		return nil == right
	}else if nil == right{
		return true
	}else{
		if len(left) < len(right){
			return false
		}
		for i, cr := range right{
			if left[i] != cr{
				return false
			}
		}
		return true
	}
}