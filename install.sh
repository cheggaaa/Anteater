#!/bin/bash

BINNAME="anteater"
GOAPP=$( cd "$( dirname "$0" )" && pwd )
GOPATH=$GOPATH:$GOAPP
GOBIN=$GOAPP/bin

echo "Install go pkgs.."
go get testing
go get github.com/akrennmair/goconf
go get github.com/cheggaaa/pb
go get github.com/valyala/fasthttp
go get github.com/valyala/fasthttp/fasthttpadaptor

do_run_test() {
	echo "Run tests.."
	cd $GOAPP/src
	GOPATH=$GOPATH go test --race utils config dump storage $@
	cd ../
}

do_run_build() {
	GOPATH=$GOPATH go install -v ./...
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
