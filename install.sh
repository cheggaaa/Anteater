#!/bin/bash

cd ./src
GOPATH=$(pwd)/src go build -v -o ../bin/aecommand cmd/aecommand/aecommand.go
GOPATH=$(pwd)/src go build -v -o ../bin/aeimport cmd/aeimport/aeimport.go
GOPATH=$(pwd)/src go build -v -o ../bin/aestats cmd/aestats/aestats.go
GOPATH=$(pwd)/src go build -v -o ../bin/anteater cmd/anteater/anteater.go
cd ../
