#!/bin/bash

BINNAME="anteater"
AEROOT=$( cd "$( dirname "$0" )" && pwd )
GOBIN=$AEROOT/bin

install -d $GOBIN
cd src/
echo "Building.."
GOBIN=$GOBIN go install
mv $GOBIN/src $GOBIN/$BINNAME
