#!/bin/bash

set -e
echo "case 1: simple get"
./httpgen -t py curl http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 2: simple post with data"
./httpgen -t py curl -d test http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 3: post multiple datas"
./httpgen -t py curl -d test -d hello http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 4: post url encoded data"
./httpgen -t py curl --data-urlencode="test% =" http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 5: get with parameter"
./httpgen -t py curl -G -d hello http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 6: get with aprameter (key=value style)"
./httpgen -t py curl -G -d hello=world http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 7: post with aprameter (key=value style)"
./httpgen -t py curl -X POST -G -d hello=world http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 8: simple post without data"
./httpgen -t py curl -X POST http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 9: simple post with local file content"
./httpgen -t py curl -X POST -T test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 10: post form"
./httpgen -t py curl -F hello=world http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 11: post text data from local file"
./httpgen -t py curl --data-ascii @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 12: post text data from local files"
./httpgen -t py curl --data-ascii @test.py --data-ascii @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 13: post data from local file"
./httpgen -t py curl --data-binary @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 14: post data from local files"
./httpgen -t py curl --data-binary @test.py --data-binary @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 15: post url encoded data from local file"
./httpgen -t py curl --data-urlencode @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 16: post url encoded data from local files"
./httpgen -t py curl --data-urlencode @test.py --data-urlencode @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 17: send file in form protocol"
./httpgen -t py curl -F "file=@test.py" http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 18: send file in form protocol with explicit name and type"
./httpgen -t py curl -F "file=@test.py;filename=nameinpost;type=text/plain" http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 19: send file in form protocol (2)"
./httpgen -t py curl -F "file=<test.py" http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 20: send file in form protocol with explicit type (2)"
./httpgen -t py curl -F "file=<test.py;type=text/plain" http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 21: get with aprameter and header"
./httpgen -t py curl -H "Accept: text/html" -G -d hello=world http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 22: simple post with data and user-agent"
./httpgen -t py curl --user-agent="Netscape 4.7" -d test http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 23: post url encoded data from local files with compressed option"
./httpgen -t py curl --compressed --data-urlencode @test.py --data-urlencode @test.py http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

echo "case 24: Basic authentication"
./httpgen -t py curl -u USER:PASS http://localhost:18888 > test/test.py
pushd test;python3 test.py;popd

