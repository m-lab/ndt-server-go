// Part of ndt-server-go <https://github.com/m-lab/ndt-server-go>, which
// is free software under the Apache v2.0 License.

package util

import (
	"testing"
	"time"
)

// TestApiWorks makes sure that we correctly deal with all the possible
// range of input parameters and produce consistent output.
func TestApiWorks(t *testing.T) {
	check := func(f func(BytesGenerator, int) []byte) {
		bgen := NewBytesGenerator()
		o := f(bgen, -128)
		if len(o) != 0 {
			t.Error("cannot deal with negative input")
		}
		o = f(bgen, 0)
		if len(o) != 0 {
			t.Error("cannot deal with zero input")
		}
		o = f(bgen, 512)
		if len(o) != 512 {
			t.Error("cannot deal with positive input")
		}
	}
	check(func(bgen BytesGenerator, n int) []byte {
		return bgen.GenLettersFast(n)
	})
	check(func(bgen BytesGenerator, n int) []byte {
		return bgen.GenAnythingSlow(n, []byte("ABCDEabcdeZz"))
	})
}

// TestIsSeeded makes sure that two subsequently created generators
// do not typically generate equal random sequences. There may possibly
// some cases in which this could fail (i.e. the clock jumping back
// because of some cloud hiccups). I added a sleep in here to reduce
// the likelyhood of that, but stil it may fail sometimes.
func TestIsSeeded(t *testing.T) {
	check := func(f func(BytesGenerator, int) []byte) {
		first := NewBytesGenerator()
		time.Sleep(100 * time.Millisecond) // Be sure
		second := NewBytesGenerator()
		if string(f(first, 1024)) == string(f(second, 1024)) {
			t.Error("seems we are not seeded")
		}
	}
	check(func(bgen BytesGenerator, n int) []byte {
		return bgen.GenLettersFast(n)
	})
	check(func(bgen BytesGenerator, n int) []byte {
		return bgen.GenAnythingSlow(n, []byte("ABCDEabcdeZz"))
	})
}
