#!/bin/sh

go-bindata -pkg="generator" -o="generator/bindata.go" templates
go fmt generator/*.go
go build
#pushd form2curl/form2curl; go build; popd
pushd webui; gopherjs build; popd
