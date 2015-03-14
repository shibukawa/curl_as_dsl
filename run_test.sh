#!/bin/bash

set +e
mkdir test

set -e

./run_test_go.sh
./run_test_py.sh
./run_test_nodejs.sh
./run_test_java.sh
./run_test_objc.sh
./run_test_objc_connection.sh
./run_test_php.sh
