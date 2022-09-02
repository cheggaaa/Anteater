#!/bin/bash

go build -v -o bin/aecommand cmd/aecommand/aecommand.go
go build -v -o bin/aeimport cmd/aeimport/aeimport.go
go build -v -o bin/aestats cmd/aestats/aestats.go
go build -v -o bin/anteater cmd/anteater/anteater.go
go build -v -o bin/aemove cmd/aemove/aemove.go
 
