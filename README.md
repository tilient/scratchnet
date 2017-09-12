scratchnet
==========

A helper application for Scratch.

It enables multiple machines to talk with each other from within Scratch.

downloads
---------

scratch extension file: [scratchnet.s2e](https://github.com/tilient/scratchnet/releases/download/v0.0-alpha/scratchnet.s2e)


linux binary (64-bit): [main.linux.64-bit](https://github.com/tilient/scratchnet/releases/download/v0.0-alpha/main.linux.64-bit)

windows binary (64-bit): [main.exe.windows.64-bit.exe](https://github.com/tilient/scratchnet/releases/download/v0.0-alpha/main.exe.windows.64-bit.exe)

to build for windows
--------------------

on linux using cross-compilation

needs _mingw_: install _gcc-mingw-w64-x86-64_

    export CGO_ENABLED=1
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    go build -ldflags "-H windowsgui" -o main.exe main.go
