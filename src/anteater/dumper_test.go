package anteater

import (
	"testing"
	"os"
)


const (
	IndexFilename = "/tmp/anteater.test.file.index"
)

func init() {
	os.Remove(IndexFilename)
	IndexSet("testDump1", &FileInfo{Id:11})
	IndexSet("testDump2", &FileInfo{Id:22})
	IndexSet("testDump3", &FileInfo{Id:33})
	Log, _ = LogInit()
	ContainerLastId = 42
}

func TestDumper(t *testing.T) {
	err := DumpData(IndexFilename)
	defer os.Remove(IndexFilename)
	if err != nil {
		t.Errorf("Dump has error: %v\n", err)
	}
	
	fi, err := os.Lstat(IndexFilename)
	if err != nil {
		t.Errorf("FileStat has error: %v\n", err)
	}
	
	if fi.Size() != IndexFileSize {
		t.Errorf("var IndexFileSize (%d) not math file size (%d)", IndexFileSize, fi.Size())
	}
	
	IndexDelete("testDump1")

	ContainerLastId = 1
	
	err = LoadData(IndexFilename)
	if err != nil {
		t.Errorf("Dump load data has error: %v\n", err)
	}
	
	i, ok := IndexGet("testDump1")
	
	if !ok {
		t.Errorf("LoadData do not restore index %s\n", "testDump1")
	}
	
	if i.Id != 11 {
		t.Errorf("Broken index! Id must be 11. Result: %d\n", i.Id)
	}
	
	if ContainerLastId != 42 {
		t.Errorf("LoadData do not restore ContainerLastId. Result: %d\n", ContainerLastId)
	}
} 