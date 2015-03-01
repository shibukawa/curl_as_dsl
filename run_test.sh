#!/bin/bash

set +e
mkdir test

set -e

./run_test_go.sh
./run_test_py.sh
./run_test_nodejs.sh
./run_test_java.sh
