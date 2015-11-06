#!/bin/bash

set -e
echo "case 1: simple get"
./httpgen -t node curl http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 2: simple post with data"
./httpgen -t node curl -d test http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 3: post multiple datas"
./httpgen -t node curl -d test -d hello http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 4: post url encoded data"
./httpgen -t node curl --data-urlencode="test% =" http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 5: get with parameter"
./httpgen -t node curl -G -d hello http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 6: get with aprameter (key=value style)"
./httpgen -t node curl -G -d hello=world http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 7: post with aprameter (key=value style)"
./httpgen -t node curl -X POST -G -d hello=world http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 8: simple post without data"
./httpgen -t node curl -X POST http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 9: simple post with local file content"
./httpgen -t node curl -X POST -T test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 10: post form"
./httpgen -t node curl -F hello=world http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 10-2: post form (2)"
./httpgen -t node curl -F hello=world -F good=morning http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 11: post text data from local file"
./httpgen -t node curl --data-ascii @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 12: post text data from local files"
./httpgen -t node curl --data-ascii @test.js --data-ascii @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 13: post data from local file"
./httpgen -t node curl --data-binary @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 14: post data from local files"
./httpgen -t node curl --data-binary @test.js --data-binary @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 15: post url encoded data from local file"
./httpgen -t node curl --data-urlencode @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 16: post url encoded data from local files"
./httpgen -t node curl --data-urlencode @test.js --data-urlencode @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 17: send file in form protocol"
./httpgen -t node curl -F "file=@test.js" http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 18: send file in form protocol with explicit name and type"
./httpgen -t node curl -F "file=@test.js;filename=nameinpost;type=text/plain" http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 19: send file in form protocol (2)"
./httpgen -t node curl -F "file=<test.js" http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 20: send file in form protocol with explicit type (2)"
./httpgen -t node curl -F "file=<test.js;type=text/plain" http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 21: get with prameter and header"
./httpgen -t node curl -H "Accept: text/html" -G -d hello=world http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 22: simple post with data and user-agent"
./httpgen -t node curl --user-agent="Netscape 4.7" -d test http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 23: post url encoded data from local files with compressed option"
./httpgen -t node curl --compressed --data-urlencode @test.js --data-urlencode @test.js http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 24: Basic authentication"
./httpgen -t node curl -u USER:PASS http://localhost:18888 > test/test.js
pushd test;node test.js;popd

echo "case 25: simple get with https"
./httpgen -t node curl --insecure https://localhost:18889 > test/test.js
pushd test;node test.js;popd

echo "case 26: simple get with http2"
./httpgen -t node curl --insecure --http2 https://localhost:18889 > test/test.js
pushd test;node test.js;popd
