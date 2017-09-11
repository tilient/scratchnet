
to build for windows
--------------------

using cross-compilation


export CGO_ENABLED=1
export GOOS=windows
export GOARCH=amd64
export CC=x86_64-w64-mingw32-gcc
g obuild -ldflags "-H windowsgui"  -o main.exe main.go
