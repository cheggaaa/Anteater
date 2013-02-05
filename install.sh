#!/bin/bash

BINNAME="anteater"
GOPATH=$( cd "$( dirname "$0" )" && pwd )
GOBIN=$GOPATH/bin

echo "Install go pkgs.."
go get testing
go get github.com/akrennmair/goconf

rm -rf $GOPATH/pkg/*

do_run_test() {
	echo "Run tests.."
	cd $GOPATH/src
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
