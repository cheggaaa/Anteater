/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"github.com/cheggaaa/Anteater/src/aerpc"
	"github.com/cheggaaa/Anteater/src/aerpc/rpcclient"
	"github.com/cheggaaa/pb"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	rpcAddr   = flag.String("rpc", ":32000", "Rpc server addr")
	readAddr  = flag.String("r", "localhost:8083", "Read server addr")
	writeAddr = flag.String("w", "localhost:8081", "Write server addr")
	clients   = flag.Int("c", 1, "Clients count")
	method    = flag.String("m", "POST", "http write method")

	readUrl, writeUrl string
)

func main() {
	flag.Parse()

	aerpc.RegisterCommands()

	cmd, ok := aerpc.Commands["FILELIST"]
	if !ok {
		fmt.Printf("Call to undefined command: %s\n", "FILELIST")
		return
	}

	*method = strings.ToUpper(*method)

	switch *method {
	case "PUT":
	case "POST":
	case "DELETE":
		break
	default:
		fmt.Println("Unexpected http method:", *method)
	}

	if !strings.HasPrefix(*readAddr, "http") {
		readUrl = "http://" + *readAddr + "/"
	} else {
		readUrl = *readAddr + "/"
	}
	if !strings.HasPrefix(*writeAddr, "http") {
		writeUrl = "http://" + *writeAddr + "/"
	} else {
		writeUrl = *writeAddr + "/"
	}

	addr := aerpc.NormalizeAddr(*rpcAddr)
	client, err := rpcclient.NewClient(addr)
	fmt.Println("Connect to", addr)
	if err != nil {
		fmt.Printf("Can't connect to %s:\n%v\n", addr, err)
		return
	}
	defer client.Close()

	fmt.Println("Get list of files...")
	if cmd.Execute(client); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fileList := cmd.Data().([]string)
	bar := pb.StartNew(len(fileList))
	wg := &sync.WaitGroup{}
	c := make(chan string)
	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func() {
			StartClient(c, bar)
			wg.Done()
		}()
	}

	for _, file := range fileList {
		c <- file
	}
	close(c)
	wg.Wait()
	bar.Finish()
}

func StartClient(c chan string, bar *pb.ProgressBar) {
	readClient, writeClient := http.Client{}, http.Client{}
	for file := range c {
		file = escapeFile(file)
		read := readUrl + file
		write := writeUrl + file
		// get
		resp, err := readClient.Get(read)
		if err != nil {
			panic(err)
		}
		getMd5 := resp.Header.Get("X-Ae-Md5")
		length := resp.ContentLength

		// put
		req, err := http.NewRequest(*method, write, resp.Body)
		if err != nil {
			panic(err)
		}
		req.ContentLength = length
		wres, err := writeClient.Do(req)
		if err != nil {
			panic(err)
		}
		if wres.StatusCode != http.StatusConflict {
			putMd5 := wres.Header.Get("X-Ae-Md5")
			if putMd5 != getMd5 {
				fmt.Printf("ERROR! MD5 not equals: %s vs %s (%s)\n", getMd5, putMd5, file)
			}
		}
		wres.Body.Close()
		resp.Body.Close()
		bar.Increment()
	}
}

func escapeFile(name string) (result string) {
	u := &url.URL{}
	u.Path = name
	return u.RequestURI()
}
