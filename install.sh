#!/bin/bash
#GOPATH=/home/che/projects/Anteater/ GOBIN=/home/che/projects/Anteater/bin/ go install ./src/anteater.go
BINNAME="anteater"
GOPATH=$( cd "$( dirname "$0" )" && pwd )
GOBIN=$GOPATH/bin

echo "Install go pkgs.."
go get testing
go get github.com/kless/goconfig/config

if [[ $1 != "notest" ]]; then
	echo "Run tests.."
	cd $GOPATH/src
	GOPATH=$GOPATH go test utils config dump storage
	cd ../
fi

if [[ $1 != "test" ]]; then
	install -d $GOBIN
	echo "Building anteater.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/anteater.go
	echo "Building aecommand.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/aecommand.go
	echo "Building aeimport.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/aeimport.go
fi
