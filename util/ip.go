package util

import (
	"strconv"
	"strings"
)

func ConvertIP(ip string) (b []byte) {
	as := strings.Split(ip, ".")
	if len(as) != 4 {
		return
	}
	for _, a := range as {
		n, err := strconv.Atoi(a)
		if err != nil {
			return
		}
		b = append(b, byte(n))
	}
	return
}
