scratchnet
==========

A helper application for Scratch.

It enables multiple machines to talk with each other from within Scratch.

downloads
---------

scratch extension file: xxx


linux binary (64-bit): xxx

windows binary (64-bit): xxx

to build for windows
--------------------

on linux using cross-compilation

needs _mingw_: install _gcc-mingw-w64-x86-64_

    export CGO_ENABLED=1
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    go build -ldflags "-H windowsgui" -o main.exe main.go
