#!/bin/bash -ex

export PATH="$PATH:${GOPATH%%:*}/bin"
echo Installing govendor
go get -u github.com/kardianos/govendor
echo Obtaining dependencies
govendor sync