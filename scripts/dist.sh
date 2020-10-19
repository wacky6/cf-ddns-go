#!/bin/bash

set -e

OS_ARCHS=(
    linux/amd64
    linux/arm
    darwin/amd64
    windows/amd64
)

if command -v upx &> /dev/null
then
    echo "* Found UPX. Will compress binaries."
    HAS_UPX=1
else
    HAS_UPX=0
fi

if command -v git &> /dev/null
then
    BUILD_COMMIT=$( git rev-parse --short HEAD )
    BUILD_TAG=$( git describe --exact-match --tags HEAD 2>/dev/null|| echo "" )
else
    BUILD_COMMIT="<Unknown>"
    BUILD_TAG=""
fi

BUILD_TIME=$( date -u "+%Y-%m-%d %H:%M:%S UTC" )
BUILD_INFO=" \
    -X 'main.buildTime=${BUILD_TIME}' \
    -X 'main.buildTag=${BUILD_TAG}' \
    -X 'main.buildCommit=${BUILD_COMMIT}' \
"

echo "* Build time:   $BUILD_TIME"
echo "* Build commit: $BUILD_COMMIT"
echo "* Build tag:    $BUILD_TAG"
echo ""

for OS_ARCH in "${OS_ARCHS[@]}" ; do
    IFS='/' read -r OS ARCH <<< "$OS_ARCH"

    BINARY_FILENAME="cf-ddns-${OS}-${ARCH}"
    if [ $OS == "windows" ] ; then
        BINARY_FILENAME=${BINARY_FILENAME}.exe
    fi

    echo "*** Building ${OS} ${ARCH}"

    GOOS=$OS GOARCH=$ARCH \
        go build \
        -ldflags="-s -w $BUILD_INFO" \
        -o ./bin/${BINARY_FILENAME} \
        ./cmd

    if [[ $HAS_UPX -eq 1 ]] ; then
        upx -qqq ./bin/${BINARY_FILENAME}
    fi

    FILE_SIZE=$( du -h ./bin/${BINARY_FILENAME} | cut -f1 )
    echo "*** Built ${OS} ${ARCH}, size: ${FILE_SIZE}"
done
