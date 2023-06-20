#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
apt update
apt install -y wget xfslibs-dev libzstd-dev liblz-dev liblz4-dev liblzo2-dev zlib1g-dev liblzma-dev build-essential
NAME=snapd
BUILD_DIR=${DIR}/build/${NAME}

rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}
mkdir ${BUILD_DIR}/bin

wget https://github.com/plougher/squashfs-tools/archive/refs/tags/4.4.tar.gz
tar xf 4.4.tar.gz
cd squashfs-tools-4.4/squashfs-tools
sed -i 's/#XZ_SUPPORT.*/XZ_SUPPORT=1/g' Makefile
sed -i 's/#LZO_SUPPORT.*/LZO_SUPPORT=1/g' Makefile
sed -i 's/#LZ4_SUPPORT.*/LZ4_SUPPORT=1/g' Makefile
sed -i 's/#ZSTD_SUPPORT.*/ZSTD_SUPPORT=1/g' Makefile
LDFLAGS=-static make
cp mksquashfs ${BUILD_DIR}/bin
cp unsquashfs ${BUILD_DIR}/bin
