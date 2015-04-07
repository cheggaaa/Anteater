#!/bin/bash

BINNAME="anteater"
GOAPP=$( cd "$( dirname "$0" )" && pwd )
GOPATH=$GOPATH:$GOAPP
GOBIN=$GOAPP/bin

echo "Install go pkgs.."
go get testing
go get github.com/akrennmair/goconf
go get github.com/cheggaaa/pb

rm -rf $GOAPP/pkg/*

do_run_test() {
	echo "Run tests.."
	cd $GOAPP/src
	GOPATH=$GOPATH go test utils config dump storage $@
	cd ../
}

do_run_build() {
	install -d $GOBIN
	echo "Building anteater.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/anteater.go
	echo "Building aecommand.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/aecommand.go
	echo "Building aeimport.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/aeimport.go
	echo "Building aestats.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/aestats.go
	echo "Building aemove.."
	GOBIN=$GOBIN GOPATH=$GOPATH go install ./src/aemove.go
}

case $1 in
	test)
		shift
		do_run_test $@
		;;
	notest)
		do_run_build
		;;
	*)
		do_run_test
		do_run_build
		;;
esac
