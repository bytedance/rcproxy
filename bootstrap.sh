#!/bin/sh

make clean

echo "downloading dependcies, it may take a few minutes..."

godep path > /dev/null 2>&1
if [ "$?" = 0 ]; then
    GOPATH=`godep path`:$GOPATH
    godep restore
    exit 0
else
    echo "godep is not found, @see https://github.com/tools/godep"
    exit 1
fi
