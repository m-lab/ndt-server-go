// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"math/rand"
	"time"
)

// BytesGenerator is a random bytes generator. We use it to generate random
// sequences of characters in the NDT test. BytesGenerator implements optimized
// techniques described in https://stackoverflow.com/a/12321192. One issue of
// BytesGenerator is that it's not multi-goroutine safe. When serving a client
// with multiple goroutines, create a BytesGenerator per goroutine!
type BytesGenerator struct {
	src rand.Source
}

// NewBytesGenerator returns a BytesGenerator. It seeds the internal randomness
// source with the current time, so you do not have to worry about that.
func NewBytesGenerator() BytesGenerator {
	return BytesGenerator{
		src: rand.NewSource(time.Now().UnixNano()),
	}
}

// GenLettersFast generates a |n| sized bytes vector containing only ASCII
// uppercase and lowercase letters chosen at random. The algorithm used
// by this function is the fastest according to the above mentioned thread
// on Stack Overflow. Yet, it is not flexible in that it is specifically
// optimized for only returning a character in a 52 characters set.
func (bgen BytesGenerator) GenLettersFast(n int) []byte {
	if n <= 0 {
		return make([]byte, 0)
	}

	// WARNING: don't change this variable without changing also the
	// algorithm to work with a string of different length!
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	const (
		// 6 bits to represent a letter index
		letterIdxBits = 6
		// All 1-bits, as many as letterIdxBits
		letterIdxMask = 1 << (letterIdxBits - 1)
		// Number of letter indices fitting in 63 bits
		letterIdxMax = 63 / letterIdxBits
	)

	source := bgen.src
	b := make([]byte, n)

	// A rand.Int63() generates 63 random bits, enough for
	// letterIdxMax characters!
	for i, cache, remain := n-1, source.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = source.Int63(), letterIdxMax
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

// GenAnythingSlow generates a |n| sized bytes vector containing only chars
// that appear within |input|, chosen at random. According to the above
// mentioned Stack Overflow thread, this is one of the slowest methods to
// generate random bytes, but we include it because it's flexible.
func (bgen BytesGenerator) GenAnythingSlow(n int, input []byte) []byte {
	if n <= 0 {
		return make([]byte, 0)
	}
	source := bgen.src
	b := make([]byte, n)
	for i := range b {
		b[i] = input[source.Int63()%int64(len(input))]
	}
	return b
}
