# scratchnet

A helper application for Scratch.
It enables multiple machines to talk with each other from within Scratch.

downloads
---------

linux (64-bit): https://github.com/tilient/scratchnet/raw/master/builds/windows64/main

windows (64-bit): https://github.com/tilient/scratchnet/raw/master/builds/windows64/main.exe

to build for windows
--------------------

on linux using cross-compilation

needs _mingw_: install _gcc-mingw-w64-x86-64_

    export CGO_ENABLED=1
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    go build -ldflags "-H windowsgui" -o main.exe main.go
