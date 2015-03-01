#!/bin/bash

set -e
echo "case 1: simple get"
./httpgen -t java curl http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 2: simple post with data"
./httpgen -t java curl -d test http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 3: post multiple datas"
./httpgen -t java curl -d test -d hello http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 4: post url encoded data"
./httpgen -t java curl --data-urlencode="test% =" http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 5: get with parameter"
./httpgen -t java curl -G -d hello http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 6: get with aprameter (key=value style)"
./httpgen -t java curl -G -d hello=world http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 7: post with aprameter (key=value style)"
./httpgen -t java curl -X POST -G -d hello=world http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 8: simple post without data"
./httpgen -t java curl -X POST http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 9: simple post with local file content"
./httpgen -t java curl -X POST -T Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 10: post form"
./httpgen -t java curl -F hello=world http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 11: post text data from local file"
./httpgen -t java curl --data-ascii @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 12: post text data from local files"
./httpgen -t java curl --data-ascii @Main.java --data-ascii @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 13: post data from local file"
./httpgen -t java curl --data-binary @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 14: post data from local files"
./httpgen -t java curl --data-binary @Main.java --data-binary @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 15: post url encoded data from local file"
./httpgen -t java curl --data-urlencode @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 16: post url encoded data from local files"
./httpgen -t java curl --data-urlencode @Main.java --data-urlencode @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 17: send file in form protocol"
./httpgen -t java curl -F "file=@Main.java" http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 18: send file in form protocol with explicit name and type"
./httpgen -t java curl -F "file=@Main.java;filename=nameinpost;type=text/plain" http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 19: send file in form protocol (2)"
./httpgen -t java curl -F "file=<Main.java" http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 20: send file in form protocol with explicit type (2)"
./httpgen -t java curl -F "file=<Main.java;type=text/plain" http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 21: get with aprameter and header"
./httpgen -t java curl -H "Accept: text/html" -G -d hello=world http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 22: simple post with data and user-agent"
./httpgen -t java curl --user-agent="Netscape 4.7" -d test http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 23: post url encoded data from local files with compressed option"
./httpgen -t java curl --compressed --data-urlencode @Main.java --data-urlencode @Main.java http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

echo "case 24: Basic authentication"
./httpgen -t java curl -u USER:PASS http://localhost:18888 > test/Main.java
pushd test;javac Main.java;java Main;popd

