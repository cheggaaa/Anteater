package storage

import (
	"config"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

const (
	J_OP_DELETE = 1
	J_OP_ADD    = 2
	J_OP_RENAME = 3

	J_E_BLOCK_SIZE  = 1200
	J_E_HEADER_SIZE = 100
)

type Journal struct {
	Header  *JournalHeader
	StartId int64
	Cap     int32
	jids    []int64
	f       *os.File
	m       *sync.Mutex
	c       *config.Config
}

type JournalHeader struct {
	Size    int64
	Index   int32
	Version int64
}

type JournalEntry struct {
	// operation id
	O int
	// previous operation id
	P int
	// container id
	C int64
	// journal id
	I int64
	// file
	F *File
}

func (k *Journal) Init(c *config.Config) {

}

func (j *Journal) fileName() string {
	return fmt.Sprintf("%sjournal", j.c.DataPath)
}

func (j *Journal) NewId() int64 {
	return atomic.AddInt64(&j.Header.Version, 1)
}
