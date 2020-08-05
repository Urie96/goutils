package goutils

import (
	"math"
	"strings"
	"time"
)

func GenerateTicketNumber(operation string) string {
	prefix := strings.ToUpper(string(operation[0]))
	timeString := time.Now().UTC().Format("20060102")
	randomString := generateRandomString(5)
	return prefix + timeString + randomString
}

var generateRandomString = func() func(length int) string {
	charset := []byte("0123456789ABCDEFGHIJKLMNPQRSTUVWXYZ")
	len := len(charset)
	return func(length int) string {
		t := time.Now().UTC()
		precision := float64(3600*24*1e9) / math.Pow(float64(len), float64(length))
		num := int(float64((t.Hour()*3600+t.Minute()*60+t.Second())*1e9+t.Nanosecond()) / precision)
		res := make([]byte, length)
		i := length - 1
		for num > 0 {
			res[i] = charset[num%len]
			num /= len
			i--
		}
		for i >= 0 {
			res[i] = charset[0]
			i--
		}
		return string(res)
	}
}()
