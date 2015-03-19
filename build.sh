#!/bin/sh

go-bindata -pkg="generator" -o="generator/bindata.go" templates
go fmt generator/*.go
go build
pushd webui;gopherjs build; popd
