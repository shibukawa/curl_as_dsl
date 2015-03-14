#!/bin/bash

set -e
echo "case 1: simple get"
./httpgen -t php curl http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 2: simple post with data"
./httpgen -t php curl -d test http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 3: post multiple datas"
./httpgen -t php curl -d test -d hello http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 4: post url encoded data"
./httpgen -t php curl --data-urlencode="test% =" http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 5: get with parameter"
./httpgen -t php curl -G -d hello http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 6: get with aprameter (key=value style)"
./httpgen -t php curl -G -d hello=world http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 7: post with aprameter (key=value style)"
./httpgen -t php curl -X POST -G -d hello=world http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 8: simple post without data"
./httpgen -t php curl -X POST http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 9: simple post with local file content"
./httpgen -t php curl -X POST -T test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 10: post form"
./httpgen -t php curl -F hello=world http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 10-2: post form (2)"
./httpgen -t php curl -F hello=world -F good=morning http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 11: post text data from local file"
./httpgen -t php curl --data-ascii @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 12: post text data from local files"
./httpgen -t php curl --data-ascii @test.php --data-ascii @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 13: post data from local file"
./httpgen -t php curl --data-binary @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 14: post data from local files"
./httpgen -t php curl --data-binary @test.php --data-binary @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 15: post url encoded data from local file"
./httpgen -t php curl --data-urlencode @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 16: post url encoded data from local files"
./httpgen -t php curl --data-urlencode @test.php --data-urlencode @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 17: send file in form protocol"
./httpgen -t php curl -F "file=@test.php" http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 18: send file in form protocol with explicit name and type"
./httpgen -t php curl -F "file=@test.php;filename=nameinpost;type=text/plain" http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 19: send file in form protocol (2)"
./httpgen -t php curl -F "file=<test.php" http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 20: send file in form protocol with explicit type (2)"
./httpgen -t php curl -F "file=<test.php;type=text/plain" http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 21: get with aprameter and header"
./httpgen -t php curl -H "Accept: text/html" -G -d hello=world http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 22: simple post with data and user-agent"
./httpgen -t php curl --user-agent="Netscape 4.7" -d test http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 23: post url encoded data from local files with compressed option"
./httpgen -t php curl --compressed --data-urlencode @test.php --data-urlencode @test.php http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

echo "case 24: Basic authentication"
./httpgen -t php curl -u USER:PASS http://localhost:18888 > test/test.php
pushd test;php56 test.php;popd

