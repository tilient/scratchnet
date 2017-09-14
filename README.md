scratchnet
==========

A extension for Scratch that enables multiple machines to talk with each other from within Scratch.

downloads
---------

linux binary (64-bit): [scratchnet](https://github.com/tilient/scratchnet/releases/download/v0.3-alpha/scratchnet)

windows binary (64-bit): [scratchnet.exe](https://github.com/tilient/scratchnet/releases/download/v0.3-alpha/scratchnet.exe)

example scratch project: [project.sb2](https://github.com/tilient/scratchnet/releases/download/v0.3-alpha/project.sb2)

dependencies
------------

    go get github.com/GeertJohan/go.rice
    go get github.com/GeertJohan/go.rice/rice


    sudo apt install gcc-mingw-w64-x86-64

to build for linux
------------------

    go build
    rice append --exec scratchnet


to build for windows
--------------------

to build on linux using cross-compilation,
_mingw_ is needed. Install _gcc-mingw-w64-x86-64_
and execute the following commands.

    export CGO_ENABLED=1
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    go build -ldflags "-H windowsgui" -o scratchnet.exe
    rice append --exec scratchnet.exe
