package storage

import (
	"time"
)

type File struct {
	Name string
	CId int32
	Start int64
	Size int64
	Time time.Time
	Md5 []byte
}