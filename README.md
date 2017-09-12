# scratchnet


to build for windows
--------------------

on linux using cross-compilation

needs _mingw_: install _x86_64-w64-mingw32-gcc_

    export CGO_ENABLED=1
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    go build -ldflags "-H windowsgui" -o main.exe main.go
