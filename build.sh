#!/bin/bash

if [[ "$1" == "C" ]]; then
    go build -o build/libtau.so -buildmode=c-shared .
fi
