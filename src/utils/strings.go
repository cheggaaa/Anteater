package utils

import (
	"fmt"
	"strconv"
	"strings"
)


// Convert bytes to human readable string. Like a 2 MiB, 64.2 KiB, 52 B
func HumanBytes(size int64) (result string) {
	switch {
	case size > (1024 * 1024 * 1024 * 1024):
		result = fmt.Sprintf("%6.2f TiB", float64(size)/1024/1024/1024/1024)
	case size > (1024 * 1024 * 1024):
		result = fmt.Sprintf("%6.2f GiB", float64(size)/1024/1024/1024)
	case size > (1024 * 1024):
		result = fmt.Sprintf("%6.2f MiB", float64(size)/1024/1024)
	case size > 1024:
		result = fmt.Sprintf("%6.2f KiB", float64(size)/1024)
	default:
		result = fmt.Sprintf("%d B", size)
	}
	result = strings.Trim(result, " ")
	return
}


// Convert human readable string to bytes. 1k string -> 1024 int
func BytesFromString(s string) (int64, error) {
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