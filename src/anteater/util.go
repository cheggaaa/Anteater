package anteater

import (
	"fmt"
	"strconv"
)


func SafeDivision(a, b int64) int64 {
	if b <= 0 {
		return 0
	}
	return a / b
}

func HumanBytes(size int64) (result string) {
	switch {
		case size > (1024 * 1024 * 1024 * 1024):
			result = fmt.Sprintf("%6.2f TiB", float64(size) / 1024 / 1024 / 1024 / 1024)
		case size > (1024 * 1024 * 1024):
			result = fmt.Sprintf("%6.2f GiB", float64(size) / 1024 / 1024 / 1024)
		case size > (1024 * 1024):
			result = fmt.Sprintf("%6.2f MiB", float64(size) / 1024 / 1024)
		case size > 1024:
			result = fmt.Sprintf("%6.2f KiB", float64(size) / 1024)
		default :
			result = fmt.Sprintf("%d B", size)
	}
	return
}


func GetSizeFromString(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	var m int64 = 1

	switch s[len(s)-1] {
	case 'K', 'k':
		m = 1024
	case 'M', 'm':
		m = 1024 * 1024
	case 'G', 'g':
		m = 1024 * 1024 * 1024
	case 'T', 't':
		m = 1024 * 1024 * 1024 * 1024
	}

	if m != 1 {
		s = s[0 : len(s)-1]
	}

	res, err := strconv.ParseInt(s, 0, 64)
	
	if err != nil {
		return 0, err
	}
	
	return res * m, nil
}