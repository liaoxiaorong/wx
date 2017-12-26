package wx

import (
	"math/rand"
	"strconv"
	"time"
)

var randsrc = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const letterNumber = "1234567890"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandNumbers(n int) string {
	return randStringBytesMaskImprSrc(n, letterNumber)
}

func TimestampStr() string {
	return strconv.FormatInt(Timestamp(), 10)
}
func TimestampMicroSecond() int64 {
	return time.Now().UnixNano() / 1e6
}

func Timestamp() int64 {
	return time.Now().Unix()
}

func RandString(n int) string {
	return randStringBytesMaskImprSrc(n, letterBytes)
}

func randStringBytesMaskImprSrc(n int, letters string) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randsrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randsrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
