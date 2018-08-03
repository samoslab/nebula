#!/bin/bash -x
if [ $# != 1 ]; then
  echo "USAGE: $0 version"
  exit 1;
fi
export VERSION=$1

bld(){
	echo $1
	build_out=nebula-provider
	if test $2 = 1
	then
		build_out=nebula-provider.exe
	fi
	export $4
	go build -o $build_out main.go
	DIST_FILE=dist/nebula-provider-$VERSION-$3.tar.bz2
	if test $2 = 1
	then
		DIST_FILE=nebula-provider-$VERSION-$3.zip
		zip -r $DIST_FILE $build_out
		mv $DIST_FILE dist/
		DIST_FILE=dist/nebula-provider-$VERSION-$3.zip
	else
		tar jcvfp $DIST_FILE $build_out
	fi
	rm $build_out
	sha256sum $DIST_FILE >> $DIST_FILE-sha256sum.txt
}

if [ -d dist ]
then
	rm -rf dist/*
else
	mkdir dist
fi

unset GOARM
unset GOOS
unset GOARCH
unset CGO_ENABLED
bld "build linux amd64" 0 "linux-amd64" "CGO_ENABLED=0 GOOS=linux GOARCH=amd64"
bld "build linux 386" 0 "linux-386" "CGO_ENABLED=0 GOOS=linux GOARCH=386"
bld "build mac amd64" 0 "mac-amd64" "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64"
bld "build windows amd64" 1 "windows-amd64" "CGO_ENABLED=0 GOOS=windows GOARCH=amd64"
bld "build windows 386" 1 "windows-386" "CGO_ENABLED=0 GOOS=windows GOARCH=386"
bld "build freebsd amd64" 0 "freebsd-amd64" "CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64"
bld "build freebsd 386" 0 "freebsd-386" "CGO_ENABLED=0 GOOS=freebsd GOARCH=386"
bld "build linux arm64" 0 "linux-arm64" "CGO_ENABLED=0 GOOS=linux GOARCH=arm64"
bld "build linux armv7" 0 "linux-armv7" "CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7"
bld "build linux armv6" 0 "linux-armv6" "CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6"
bld "build linux armv5" 0 "linux-armv5" "CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5"




