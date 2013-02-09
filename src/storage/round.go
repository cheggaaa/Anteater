package storage

import (
	"math"
)

var R = Rounder{}

var (
	i1kb = int64(1024)
	i16kb = int64(1024 * 16)
	i128kb = int64(1024 * 128)
	i1mb = int64(1024 * 1024)
	i4mb = i1mb * 4
)


type Rounder struct{}

/**
 * 2,4,8,16..512,1024 - n2
 * 1k,2k,3k...15k,16k - 1k
 * 16k,32k,48k..112k,128k - 16k
 * 128k,256k,384k..3968k,4096k - 128k
 * 4m,5m,6m... - 1mb
 **/
func (r Rounder) Index(size int64) int {
	if size == 0 {
		panic("Try get index for 0")
	}
	switch {
		case size <= i1kb:
			return r.tinyIndex(size)
		case size <= i16kb:
			return r.index(size, i1kb) + 10
		case size <= i128kb:
			return r.index(size, i16kb) + 25
		case size <= i4mb:
			return r.index(size, i128kb) + 32
	}	
	return r.index(size, i1mb) + 60
}

func (r Rounder) index(size, rg int64) int {
	return int((size -1) / rg)
}

func (r Rounder) tinyIndex(size int64) int {
	return int(math.Log2(float64(size) - 0.1)) + 1
}

func (r Rounder) Size(index int) int64 {
	if index == 0 {
		panic("Try get size for 0")
	}
	
	switch {
		case index < 11:
			return r.tinySize(index)
		case index < 26:
			return r.size(index - 9, i1kb)
		case index < 33:
			return r.size(index - 24, i16kb)
		case index < 64:
			return r.size(index - 31, i128kb)
	}
	
	return r.size(index - 59, i1mb)
}

func (r Rounder) size(index int, rg int64) int64 {
	return int64(index) * rg
}

func (r Rounder) tinySize(index int) int64 {
	return int64(math.Pow(2, float64(index)))
}

func (r Rounder) Round(size int64) int64 {
	return r.Size(r.Index(size))
}