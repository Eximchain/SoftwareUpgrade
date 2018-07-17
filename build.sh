#!/bin/bash

. ./vars.sh
go get -u github.com/kardianos/govendor

cd "$GOPATH"/src/softwareupgrade
"$GOPATH"/bin/govendor sync
cd $GOPATH
go build -o Upgrade src/LaunchUpgrade/main.go
