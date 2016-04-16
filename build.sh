#!/bin/bash

VER=$1

GOOS=linux GOARCH=386 go build -o bin/mosesacs mosesacs.go 
cd bin
tar cvfz mosesacs_${VER}_linux_i386.tar.gz mosesacs
rm mosesacs
s3cmd -c ~/.s3cfg put mosesacs_${VER}_linux_i386.tar.gz s3://mosesacs/
cd ..

GOOS=linux GOARCH=amd64 go build -o bin/mosesacs mosesacs.go
cd bin
tar cvfz mosesacs_${VER}_linux_amd64.tar.gz mosesacs
rm mosesacs
s3cmd -c ~/.s3cfg put mosesacs_${VER}_linux_amd64.tar.gz s3://mosesacs/
cd ..

GOOS=linux GOARCH=arm go build -o bin/mosesacs mosesacs.go
cd bin
tar cvfz mosesacs_${VER}_arm.tar.gz mosesacs
rm mosesacs
s3cmd -c ~/.s3cfg put mosesacs_${VER}_arm.tar.gz s3://mosesacs/
cd ..

GOOS=darwin GOARCH=amd64 go build -o bin/mosesacs mosesacs.go
cd bin
tar cvfz mosesacs_${VER}_osx_amd64.tar.gz mosesacs
rm mosesacs
s3cmd -c ~/.s3cfg put mosesacs_${VER}_osx_amd64.tar.gz s3://mosesacs/
cd ..

GOOS=windows GOARCH=386 go build -o bin/mosesacs.exe mosesacs.go
cd bin
tar cvfz mosesacs_${VER}_win_i386.tar.gz mosesacs.exe
rm mosesacs.exe
s3cmd -c ~/.s3cfg put mosesacs_${VER}_win_i386.tar.gz s3://mosesacs/
cd ..

GOOS=windows GOARCH=amd64 go build -o bin/mosesacs.exe mosesacs.go
cd bin
tar cvfz mosesacs_${VER}_win_amd64.tar.gz mosesacs.exe
rm mosesacs.exe
s3cmd -c ~/.s3cfg put mosesacs_${VER}_win_amd64.tar.gz s3://mosesacs/
cd ..
