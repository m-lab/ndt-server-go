// Part of Neubot <https://neubot.nexacenter.org/>.
// Neubot is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

// Written by Simone Basso and Davide Allavena.

package util

import (
	"math/rand"
	"time"
)

// The following is what is called "Masking Improved"
// by https://stackoverflow.com/questions/22892120

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func InitRng() {
	// Make sure we seed the random number generator properly
	//   see <http://stackoverflow.com/a/12321192>
	rand.Seed(time.Now().UTC().UnixNano())
}

func RandByteMaskingImproved(n int) []byte {
	b := make([]byte, n)

	// A rand.Int63() generates 63 random bits, enough for
	// letterIdxMax characters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
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

	return b
}

// The following is what is called "Remainder"
// by https://stackoverflow.com/questions/22892120

const rand_ascii = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

func RandAsciiRemainder(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = rand_ascii[rand.Int63()%int64(len(rand_ascii))]
	}
	return b
}
