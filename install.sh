#!/bin/bash

BINNAME="anteater"
AEROOT=$( cd "$( dirname "$0" )" && pwd )
GOBIN=$AEROOT/bin

echo "Install go pkgs.."
go get testing
go get github.com/kless/goconfig/config

echo "Run tests.."
cd src/anteater && go test && cd ../../

if [[ $1 != "test" ]]; then
	install -d $GOBIN
	echo "Building anteater.."
	GOBIN=$GOBIN go install ./src/anteater.go
	echo "Building aecommand.."
	GOBIN=$GOBIN go install ./src/aecommand.go
	echo "Building aeimport.."
        GOBIN=$GOBIN go install ./src/aeimport.go
fi
