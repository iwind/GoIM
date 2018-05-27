package stringutil

import (
	"regexp"
	"crypto/md5"
	"fmt"
	"time"
	"math/rand"
)

var reuseRegexpMap = map[string]*regexp.Regexp{}

// 判断slice中是否包含某个字符串
func Contains(slice []string, item string) bool {
	for _, _item := range slice {
		if _item == item {
			return true
		}
	}
	return false
}

// 生成可重用的正则
func RegexpCompile(pattern string) (*regexp.Regexp, error) {
	reg, ok := reuseRegexpMap[pattern]
	if ok {
		return reg, nil
	}

	reg, err := regexp.Compile(pattern)
	if err == nil {
		reuseRegexpMap[pattern] = reg
	}
	return reg, err
}

// md5
func Md5(source string) string {
	hash := md5.New()
	hash.Write([]byte(source))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// 取得随机字符串
// 代码来自 https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func Rand(n int) string {
	const randomLetterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	const (
		randomLetterIdxBits = 6                          // 6 bits to represent a letter index
		randomLetterIdxMask = 1<<randomLetterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		randomLetterIdxMax  = 63 / randomLetterIdxBits   // # of letter indices fitting in 63 bits
	)

	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), randomLetterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), randomLetterIdxMax
		}
		if idx := int(cache & randomLetterIdxMask); idx < len(randomLetterBytes) {
			b[i] = randomLetterBytes[idx]
			i--
		}
		cache >>= randomLetterIdxBits
		remain--
	}

	return string(b)
}

// 转换数字ID到字符串
func ConvertID(intId int64) string {
	const mapping = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code := ""
	size := int64(len(mapping))
	for intId >= size {
		mod := intId % size
		intId = intId / size

		code += mapping[mod:mod+1]
	}
	code += mapping[intId:intId+1]
	code = Reverse(code)

	return code
}

// 翻转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ReplaceCommentsInJSON(jsonBytes []byte) []byte {
	commentReg, err := RegexpCompile("/([*]+((.|\n|\r)+?)[*]+/)|(\n\\s+//.+)")
	if err != nil {
		panic(err)
	}
	return commentReg.ReplaceAll(jsonBytes, []byte{})
}
