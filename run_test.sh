#!/bin/bash

set +e
mkdir test

set -e
echo "case 1: simple get"
./httpgen curl http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 2: simple post with data"
./httpgen curl -d test http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 3: post multiple datas"
./httpgen curl -d test -d hello http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 4: post url encoded data"
./httpgen curl --data-urlencode="test% =" http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 5: get with parameter"
./httpgen curl -G -d hello http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 6: get with aprameter (key=value style)"
./httpgen curl -G -d hello=world http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 7: post with aprameter (key=value style)"
./httpgen curl -X POST -G -d hello=world http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 8: simple post with data"
./httpgen curl -X POST http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 8: simple post with local file content"
./httpgen curl -X POST -T test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 9: post form"
./httpgen curl -F hello=world http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 11: post text data from local file"
./httpgen curl --data-ascii @test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 12: post text data from local files"
./httpgen curl --data-ascii @test.go --data-ascii @test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 13: post data from local file"
./httpgen curl --data-binary @test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 14: post data from local files"
./httpgen curl --data-binary @test.go --data-binary @test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 15: post url encoded data from local file"
./httpgen curl --data-urlencode @test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 16: post url encoded data from local files"
./httpgen curl --data-urlencode @test.go --data-urlencode @test.go http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 17: send file in form protocol"
./httpgen curl -F "file=@test.go" http://localhost:18888 > test/test.go
pushd test;go build;./test;popd

echo "case 18: send file in form protocol with explicit name and type"
./httpgen curl -F "file=@test.go;filename=nameinpost;type=text/plain" http://localhost:18888 > test/test.go
pushd test;go build;./test;popd



