#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
apt update
apt install -y wget xfslibs-dev libzstd-dev liblz-dev liblz4-dev liblzo2-dev zlib1g-dev build-essential liblzma-dev

NAME=snapd
BUILD_DIR=${DIR}/build/${NAME}
#XZ_VERSION=5.6.4
VERSION=4.5.1

rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}
mkdir ${BUILD_DIR}/bin

#wget https://github.com/tukaani-project/xz/releases/download/v$XZ_VERSION/xz-$XZ_VERSION.tar.gz
#tar xf xz-$XZ_VERSION.tar.gz
#cd xz-$XZ_VERSION
#./configure --disable-shared
#make
#make install
#cd ..

wget https://github.com/plougher/squashfs-tools/archive/refs/tags/$VERSION.tar.gz
tar xf $VERSION.tar.gz
cd squashfs-tools-$VERSION/squashfs-tools
sed -i 's/#XZ_SUPPORT.*/XZ_SUPPORT=1/g' Makefile
sed -i 's/#LZO_SUPPORT.*/LZO_SUPPORT=1/g' Makefile
sed -i 's/#LZ4_SUPPORT.*/LZ4_SUPPORT=1/g' Makefile
sed -i 's/#ZSTD_SUPPORT.*/ZSTD_SUPPORT=1/g' Makefile
#LDFLAGS=-static
make
ldd unsquashfs || true
mkdir ${BUILD_DIR}/squashfs
mkdir ${BUILD_DIR}/squashfs/bin
cp -r $DIR/.syncloud/bin/* ${BUILD_DIR}/bin
cp mksquashfs ${BUILD_DIR}/squashfs/bin
cp unsquashfs ${BUILD_DIR}/squashfs/bin
mkdir ${BUILD_DIR}/squashfs/lib
cp ${DIR}/lib/*/ld-*.so* ${BUILD_DIR}/squashfs/lib/ld.so
cp /lib/x86_64-linux-gnu/libm.so.6 ${BUILD_DIR}/squashfs/lib
cp /lib/x86_64-linux-gnu/libz.so.1 ${BUILD_DIR}/squashfs/lib
cp /lib/x86_64-linux-gnu/liblzma.so.5 ${BUILD_DIR}/squashfs/lib
cp /lib/x86_64-linux-gnu/liblzo2.so.2 ${BUILD_DIR}/squashfs/lib
cp /lib/x86_64-linux-gnu/liblz4.so.1 ${BUILD_DIR}/squashfs/lib
cp /lib/x86_64-linux-gnu/libzstd.so.1 ${BUILD_DIR}/squashfs/lib
cp /lib/x86_64-linux-gnu/libc.so.6 ${BUILD_DIR}/squashfs/lib
