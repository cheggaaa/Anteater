#!/bin/bash

BINNAME="anteater"
AEROOT=$( cd "$( dirname "$0" )" && pwd )
GOBIN=$AEROOT/bin

echo "Run tests.."
cd src/anteater && go test && cd ../../

if [[ $1 != "test" ]]; then
	install -d $GOBIN
	echo "Building.."
	GOBIN=$GOBIN go install ./src/main.go
	mv $GOBIN/main $GOBIN/$BINNAME
fi
