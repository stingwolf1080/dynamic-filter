package types

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unsafe"
)



var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(NumStr int) string {
	b := make([]rune, NumStr)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandStringBytes(NumStr int) string {
	b := make([]byte, NumStr)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandStringBytesRmndr(NumStr int) string {
	b := make([]byte, NumStr)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func RandStringBytesMask(NumStr int) string {
	b := make([]byte, NumStr)
	for i := 0; i < NumStr; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func RandStringBytesMaskImpr(NumStr int) string {
	b := make([]byte, NumStr)
	for i, cache, remain := NumStr-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(NumStr int) string {
	b := make([]byte, NumStr)
	for i, cache, remain := NumStr-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func RandStringBytesMaskImprSrcSB(NumStr int) string {
	sb := strings.Builder{}
	sb.Grow(NumStr)
	for i, cache, remain := NumStr-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func RandStringBytesMaskImprSrcUnsafe(NumStr int) string {
	b := make([]byte, NumStr)
	for i, cache, remain := NumStr-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

type RGBColor struct {
	Red   int
	Green int
	Blue  int
}

func GetRandomColorInRgb() RGBColor {
	Red := rand.Intn(255)
	Green := rand.Intn(255)
	blue := rand.Intn(255)
	c := RGBColor{Red, Green, blue}
	return c
}

func GetRandomColorInHex() string {
	color := GetRandomColorInRgb()
	prefix := RandStringRunes(2)
	hex := prefix + getHex(color.Red) + getHex(color.Green) + getHex(color.Blue)
	return hex
}

func getHex(num int) string {
	hex := fmt.Sprintf("%x", num)
	if len(hex) == 1 {
		hex = "0" + hex
	}
	return hex
}

func GenerateTaskCode() string {
	return fmt.Sprintf("%s%04d%04d", time.Now().Format("20060102"), instanceIDToNumber(GetRandomColorInHex()), instanceIDToNumber(GetRandomColorInHex()+os.Getenv("HOSTNAME")))
}

func instanceIDToNumber(id string) int {
	h := sha1.Sum([]byte(id))
	// Lấy 2 bytes đầu tiên từ hash để tạo số 4 chữ số
	return int(h[0])<<8 + int(h[1])%1000
}

func GetRandomUserAdmin(prefix string, number int) string {
	return fmt.Sprintf("%s%s", prefix, RandStringIntegerRunes(number))
}

func RandStringIntegerRunes(NumStr int) string {
	b := make([]rune, NumStr)
	for i := range b {
		b[i] = intLetterRunes[rand.Intn(len(intLetterRunes))]
	}
	return string(b)
}

var intLetterRunes = []rune("0123456789")
