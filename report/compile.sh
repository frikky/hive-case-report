#!/bin/sh

#CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ 
#GOOS=windows GOARCH=386 CGO_ENABLED=1 go build -o run.exe gui.go cases.go

#env GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc go build -v -o main.exe 
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -v -o main.exe 
