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
    echo "***** Found UPX. Will compress binaries."
    HAS_UPX=1
else
    HAS_UPX=0
fi

for OS_ARCH in "${OS_ARCHS[@]}" ; do
    IFS='/' read -r OS ARCH <<< "$OS_ARCH"

    BINARY_FILENAME="cf-ddns-${OS}-${ARCH}"
    if [ $OS == "windows" ] ; then
        BINARY_FILENAME=${BINARY_FILENAME}.exe
    fi

    echo "***** Building ${OS} ${ARCH}"

    GOOS=$OS GOARCH=$ARCH \
        go build \
        -ldflags="-s -w" \
        -o ./bin/${BINARY_FILENAME} \
        ./cmd

    if [[ $HAS_UPX -eq 1 ]] ; then
        upx -qqq ./bin/${BINARY_FILENAME}
    fi

    FILE_SIZE=$( du -h ./bin/${BINARY_FILENAME} | cut -f1 )
    echo "***** Built ${OS} ${ARCH}, size: ${FILE_SIZE}"
done