package anteater

import (
	"encoding/gob"
	"os"
	"bytes"
	"time"
)

type Data struct {
	ContainerLastId int32
	Containers []*ContainerDumpData
	Index      map[string]*FileInfo
}

func DumpData(filename string) error {
	Log.Debugln("Dump. Start dump data...")
	
	// Timer
	tm := time.Now().UnixNano()
	
	cs := []*ContainerDumpData{}
	for _, c := range(FileContainers) {
		cs = append(cs, c.GetDumpData())
	}
	IndexLock.Lock()
	d := &Data{ContainerLastId, cs, Index}
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	Log.Debugln("Dump. Encoder created... Start encode")
	err := enc.Encode(d)
	IndexLock.Unlock()
	if err != nil {
		return err
	}
	
	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	defer fh.Close()
	if err != nil {
		return err
	}
	Log.Debugln("Dump. Write to", filename)
	n, err := fh.Write(b.Bytes())
	if err != nil {
		return err
	}
	LastDump = time.Now()
	LastDumpTime = 	(time.Now().UnixNano() - tm) / 1000
	Log.Debugf("Dump. %d bytes successfully written to file\n", n)
	return nil
}

func LoadData(filename string) error {
	Log.Debugln("Try load index from", filename);
	
	fh, err := os.Open(filename)
    if err != nil {
    	return err
    }
    d := Data{}
    dec := gob.NewDecoder(fh)
    err = dec.Decode(&d)
    if err != nil {
    	return err
    }
    
    ContainerLastId = d.ContainerLastId
    
    for _, cd := range(d.Containers) {
    	c, err := ContainerFromData(cd)
    	if err != nil {
    		return err
    	}
    	FileContainers[c.Id] = c
    }
    
    Index = d.Index
    Log.Debugln("Index loaded")
    return nil
}
