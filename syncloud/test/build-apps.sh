#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
$DIR/build.sh testapp1 1
$DIR/build.sh testapp1 2
$DIR/build.sh testapp2 1
$DIR/build.sh testapp2 2
