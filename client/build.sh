#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

bld(){
	echo $1
	build_out=nebula-client
	if test $2 = 1
	then
		build_out=nebula-client.exe
	fi
	mkdir -p $DIR/dist/$3
	export $4
	go build -o $DIR/dist/$3/$build_out $DIR/main.go
}

if [ -d $DIR/dist ]
then
	rm -rf $DIR/dist/*
else
	mkdir -p $DIR/dist
fi

unset GOARM
unset GOOS
unset GOARCH
unset CGO_ENABLED
bld "build linux amd64" 0 "linux-x64" "CGO_ENABLED=0 GOOS=linux GOARCH=amd64"
bld "build linux 386" 0 "linux-ia32" "CGO_ENABLED=0 GOOS=linux GOARCH=386"
bld "build mac amd64" 0 "mac-x64" "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64"
bld "build windows amd64" 1 "win-x64" "CGO_ENABLED=0 GOOS=windows GOARCH=amd64"
bld "build windows 386" 1 "win-ia32" "CGO_ENABLED=0 GOOS=windows GOARCH=386"
unset GOARM
unset GOOS
unset GOARCH
unset CGO_ENABLED




