package storage

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"testing"
	"time"
)

func TestReadSpaceFrom(t *testing.T) {
	var buf = bytes.NewBuffer(nil)
	hsh := md5.Sum([]byte("data"))
	f := &File{
		Hole: Hole{
			Off:  42424242,
			Indx: 56,
		},
		Name:  "some name",
		Md5:   hsh[:],
		FSize: 12345679,
		Time:  time.Now(),
	}
	err := f.MarshalTo(buf)
	if err != nil {
		t.Errorf("file marshall %v", err)
		return
	}
	f2 := &File{
		Hole: Hole{
			Off:  0,
			Indx: 873,
		},
		Name:  "second name",
		Md5:   hsh[:],
		FSize: 9979342798,
		Time:  time.Now().Add(time.Minute),
	}
	err = f2.MarshalTo(buf)
	if err != nil {
		t.Errorf("file marshall %v", err)
		return
	}
	h3 := &Hole{
		Off:  434343,
		Indx: 670,
	}
	err = h3.MarshalTo(buf)
	if err != nil {
		t.Errorf("file marshall %v", err)
		return
	}
	rd := bufio.NewReader(buf)
	s, err := UnmarshallSpace(rd)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	t.Log(s)
	s, err = UnmarshallSpace(rd)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	t.Log(s)
	s, err = UnmarshallSpace(rd)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	t.Log(s)
}
