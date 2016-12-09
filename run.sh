#!/bin/sh
go run `ls *.go|grep _test.go -v` $@