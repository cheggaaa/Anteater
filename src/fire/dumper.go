package fire

import (
	"encoding/gob"
	"os"
	"fmt"
	"bytes"
)

type Data struct {
	ContainerLastId int
	Containers []ContainerDumpData
	Index      map[string]*FileInfo
}

var DumperRunning bool

func DumpData(filename string) error {
	fmt.Println("Dump. Start dump data...")
	
	cs := []ContainerDumpData{}
	for _, c := range(FileContainers) {
		cs = append(cs, c.GetDumpData())
	}
	d := &Data{ContainerLastId, cs, Index}
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	fmt.Println("Dump. Encoder created... Start encode")
	err := enc.Encode(d)
	if err != nil {
		return err
	}
	
	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	defer fh.Close()
	if err != nil {
		return err
	}
	fmt.Println("Dump. Write to", filename)
	n, err := fh.Write(b.Bytes())
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Dump. %d bytes successfully written to file\n", n)
	
	return nil
}

func LoadData(filename string) error {
	fmt.Println("Try load index from", filename);
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
    	c, err := ContainerFromData(&cd)
    	if err != nil {
    		return err
    	}
    	FileContainers[c.Id] = c
    }
    
    Index = d.Index
    fmt.Println("Index loaded")
    return nil
}
