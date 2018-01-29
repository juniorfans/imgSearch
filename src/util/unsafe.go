package fileUtil

import "unsafe"

func GetKeyAsBytes(key *string) []byte {
	return *(* []byte)(unsafe.Pointer(key))
}

func GetKeyAsSringPtr(key *[]byte) *string {
	return (* string)(unsafe.Pointer(key))
}