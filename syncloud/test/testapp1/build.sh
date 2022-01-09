#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
ARCH=$(uname -m)
BUILD_DIR=${DIR}/build
rm -rf ${BUILD_DIR}
mkdir ${BUILD_DIR}

ARCH=$(dpkg --print-architecture)
cp -r ${DIR}/meta ${BUILD_DIR}
cp -r ${DIR}/bin ${BUILD_DIR}
echo "architectures:" >> ${BUILD_DIR}/meta/snap.yaml
echo "- ${ARCH}" >> ${BUILD_DIR}/meta/snap.yaml

mksquashfs ${BUILD_DIR} ${DIR}/testapp1.snap -noappend -comp xz -no-xattrs -all-root
cp ${DIR}/*.snap ${DIR}/../../artifact