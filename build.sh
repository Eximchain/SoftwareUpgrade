#!/bin/bash

source ./vars.sh
mkdir -p $GOPATH/bin
go get -u github.com/kardianos/govendor

cd $GOPATH/src/softwareupgrade
$GOPATH/bin/govendor sync
cd $GOPATH
go build -o Upgrade LaunchUpgrade
