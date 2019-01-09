package util

func Contains(arr []string, el string) bool {
	for _, cur := range arr {
		if el == cur {
			return true
		}
	}
	return false
}
