#!/bin/bash

set -e
echo "case 1: simple get"
./httpgen -t vim curl http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 2: simple post with data"
./httpgen -t vim curl -d test http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 3: post multiple datas"
./httpgen -t vim curl -d test -d hello http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 4: post url encoded data"
./httpgen -t vim curl --data-urlencode="test% =" http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 5: get with parameter"
./httpgen -t vim curl -G -d hello http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 6: get with aprameter (key=value style)"
./httpgen -t vim curl -G -d hello=world http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 7: post with aprameter (key=value style)"
./httpgen -t vim curl -X POST -G -d hello=world http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 8: simple post without data"
./httpgen -t vim curl -X POST http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 9: simple post with local file content"
./httpgen -t vim curl -X POST -T test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 10: post form"
./httpgen -t vim curl -F hello=world http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 10-2: post form (2)"
./httpgen -t vim curl -F hello=world -F good=morning http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 11: post text data from local file"
./httpgen -t vim curl --data-ascii @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 12: post text data from local files"
./httpgen -t vim curl --data-ascii @test.vim --data-ascii @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 13: post data from local file"
./httpgen -t vim curl --data-binary @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 14: post data from local files"
./httpgen -t vim curl --data-binary @test.vim --data-binary @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 15: post url encoded data from local file"
./httpgen -t vim curl --data-urlencode @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 16: post url encoded data from local files"
./httpgen -t vim curl --data-urlencode @test.vim --data-urlencode @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 17: send file in form protocol"
./httpgen -t vim curl -F "file=@test.vim" http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 18: send file in form protocol with explicit name and type"
./httpgen -t vim curl -F "file=@test.vim;filename=nameinpost;type=text/plain" http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 19: send file in form protocol (2)"
./httpgen -t vim curl -F "file=<test.vim" http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 20: send file in form protocol with explicit type (2)"
./httpgen -t vim curl -F "file=<test.vim;type=text/plain" http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 21: get with aprameter and header"
./httpgen -t vim curl -H "Accept: text/html" -G -d hello=world http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 22: simple post with data and user-agent"
./httpgen -t vim curl --user-agent="Netscape 4.7" -d test http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 23: post url encoded data from local files with compressed option"
./httpgen -t vim curl --compressed --data-urlencode @test.vim --data-urlencode @test.vim http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

echo "case 24: Basic authentication"
./httpgen -t vim curl -u USER:PASS http://localhost:18888 > test/test.vim
pushd test;vim -S test.vim;popd

