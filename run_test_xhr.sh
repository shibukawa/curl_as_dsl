#!/bin/bash

set -e

echo "case 1: simple get"
./httpgen -t xhr curl http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case1;sleep 1

echo "case 2: simple post with data"
./httpgen -t xhr curl -d test http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case2;sleep 1

echo "case 3: post multiple datas"
./httpgen -t xhr curl -d test -d hello http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case3;sleep 1

echo "case 4: post url encoded data"
./httpgen -t xhr curl --data-urlencode="test% =" http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case4;sleep 1

echo "case 5: get with parameter"
./httpgen -t xhr curl -G -d hello http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case5;sleep 1

echo "case 6: get with aprameter (key=value style)"
./httpgen -t xhr curl -G -d hello=world http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case6;sleep 1

echo "case 7: post with aprameter (key=value style)"
./httpgen -t xhr curl -X POST -G -d hello=world http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case7;sleep 1

echo "case 8: simple post without data"
./httpgen -t xhr curl -X POST http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case8;sleep 1

echo "case 9: simple post with local file content"
./httpgen -t xhr curl -X POST -T test.html http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case9;sleep 1

echo "case 10: post form"
./httpgen -t xhr curl -F hello=world http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case10;sleep 1

echo "case 10-2: post form (2)"
./httpgen -t xhr curl -F hello=world -F good=morning http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case10_2;sleep 1

echo "case 11: post text data from local file"
./httpgen -t xhr curl --data-ascii @test.html http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case11;sleep 1

# doesn't support it
#echo "case 12: post text data from local files"
#./httpgen -t xhr curl --data-ascii @test.html --data-ascii @test.html http://localhost:18888 > testserver/test.html
#open http://localhost:18888/js?case12;sleep 1

echo "case 13: post data from local file"
./httpgen -t xhr curl --data-binary @test.html http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case13;sleep 1

# doesn't support it
#echo "case 14: post data from local files"
#./httpgen -t xhr curl --data-binary @test.html --data-binary @test.html http://localhost:18888 > testserver/test.html
#open http://localhost:18888/js?case14;sleep 1

echo "case 15: post url encoded data from local file"
./httpgen -t xhr curl --data-urlencode @test.html http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case15;sleep 1

# doesn't support it
#echo "case 16: post url encoded data from local files"
#./httpgen -t xhr curl --data-urlencode @test.html --data-urlencode @test.html http://localhost:18888 > testserver/test.html
#open http://localhost:18888/js?case16;sleep 1

echo "case 17: send file in form protocol"
./httpgen -t xhr curl -F "file=@test.html" http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case17;sleep 1

echo "case 18: send file in form protocol with explicit name and type"
./httpgen -t xhr curl -F "file=@test.html;filename=nameinpost" http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case18;sleep 1

echo "case 19: send file in form protocol (2)"
./httpgen -t xhr curl -F "file=<test.html" http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case19;sleep 1

# doesn't support it
#echo "case 20: send file in form protocol with explicit type (2)"
#./httpgen -t xhr curl -F "file=<test.html;type=text/plain" http://localhost:18888 > testserver/test.html
#open http://localhost:18888/js?case20;sleep 1

echo "case 21: get with prameter and header"
./httpgen -t xhr curl -H "Accept: text/html" -G -d hello=world http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case21;sleep 1

echo "case 22: simple post with data and user-agent"
./httpgen -t xhr curl --user-agent="Netscape 4.7" -d test http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case22;sleep 1

echo "case 23: post url encoded data from local files with compressed option"
./httpgen -t xhr curl --compressed --data-urlencode @test.html --data-urlencode @test.html http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case23;sleep 1

echo "case 24: Basic authentication"
./httpgen -t xhr curl -u USER:PASS http://localhost:18888 > testserver/test.html
open http://localhost:18888/js?case24;sleep 1

