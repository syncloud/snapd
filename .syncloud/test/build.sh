#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
BUILD_DIR=${DIR}/build
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}
cd $DIR
go build -o $BUILD_DIR/store ./cmd/store
go build -o $BUILD_DIR/syncloud-release ./cmd/release
