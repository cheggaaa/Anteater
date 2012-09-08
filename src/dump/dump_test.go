package dump

import (
	"testing"
	"os"
	"fmt"
)

type TD struct {
	Td1 []*TD2
	Td3 *TD3
	Ti  int64
}


type TD2 struct {
	TInt int
	TString string
	TBool bool
	Td3 *TD3 
}

type TD3 struct {
	Test bool
	Test2 int64
}

var TestData *TD
var TestCount int = 1000

func TestDump(t *testing.T) {
	makeTestData()
	file := "test.dump"
	defer os.Remove(file)
	n, err := DumpTo(file, TestData)
	if err != nil {
		t.Errorf("Dump has error: %v", err)
	}
	if n <= 0 {
		t.Errorf("Dump write %d bytes. Wrong.", n)
	}
	
	data := new(TD)
	err = LoadData(file, data)
	if err != nil {
		t.Errorf("LoadData has error: %v", err)
	}
	
	if ! data.Assert() {
		t.Errorf("Data mismatched")
	}
}

func makeTestData() {
	td1 := make([]*TD2, TestCount)
	for i := 0; i < TestCount; i++ {
		td1[i] = &TD2{i, fmt.Sprintf("D:%d", i * 5), (i % 2) == 0, &TD3{(i % 3) == 0, int64(i * 5)}}
	}
	TestData = &TD{td1, &TD3{}, 12345}
}

func (t *TD) Assert() bool {
	if len(t.Td1) != TestCount {
		return false
	}
	for _, v := range t.Td1 {
		i := v.TInt
		if v.TString != fmt.Sprintf("D:%d", i * 5) {
			return false
		}
		if v.TBool != ((i % 2) == 0) {
			return false
		}
		if v.Td3.Test2 != int64(i * 5) {
			return false
		}
	}
	return true
}


