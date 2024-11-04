package utils

// 참이면 1, 거짓이면 0 반환
func ToUint(b bool) uint {
	if b {
		return 1
	}
	return 0
}
