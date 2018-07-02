#!/bin/bash

GOPATH=/Users/chuacw/Documents/GitHub/SoftwareUpgrade
go get -u -v github.com/kardianos/govendor

cd "$GOPATH"/src/SoftwareUpgrade
$GOPATH/bin/govendor sync
cd $GOPATH
go build -o Upgrade src/LaunchUpgrade/main.go
