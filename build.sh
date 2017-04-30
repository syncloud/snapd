#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [[ -z "$1" ]]; then
    echo "usage $0 version"
    exit 1
fi

VERSION=$1
TESTS=$2

export GOPATH=$( cd "$( dirname "${DIR}/../../../../.." )" && pwd )
export PATH=${PATH}:${GOPATH}/bin
NAME=snapd
BUILD_DIR=${GOPATH}/build/${NAME}
ARCH=$(dpkg-architecture -q DEB_HOST_ARCH)

cd ${GOPATH}

if [ ! -d "src/github.com/snapcore/snapd" ]; then
  echo "should be inside go path, src/github.com/snapcore/snapd"
  exit 1
fi

go get -d -v github.com/snapcore/snapd/...
cd src/github.com/snapcore/snapd

./mkversion.sh ${VERSION}

go get -u github.com/kardianos/govendor
govendor sync
if [[ ${TESTS} != "skip-tests" ]]; then
    ./run-checks
fi

cd ${GOPATH}
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

mkdir ${BUILD_DIR}/bin
go build -o ${BUILD_DIR}/bin/snapd github.com/snapcore/snapd/cmd/snapd
go build -o ${BUILD_DIR}/bin/snap github.com/snapcore/snapd/cmd/snap
go build -o ${BUILD_DIR}/bin/snap-exec github.com/snapcore/snapd/cmd/snap-exec
go build -o ${BUILD_DIR}/bin/snapctl github.com/snapcore/snapd/cmd/snapctl

cd  ${DIR}/cmd
autoreconf -i -f
./configure --disable-apparmor
make
cp snap-confine/snap-confine ${BUILD_DIR}/bin/snap-confine
cp snap-confine/snap-discard-ns ${BUILD_DIR}/bin/snap-discard-ns

cp /usr/bin/mksquashfs ${BUILD_DIR}/bin
cp /usr/bin/unsquashfs ${BUILD_DIR}/bin

mkdir ${BUILD_DIR}/lib
cp -r /lib/$(dpkg-architecture -q DEB_HOST_GNU_TYPE)/liblzo2.so* ${BUILD_DIR}/lib
cp -r /usr/lib/$(dpkg-architecture -q DEB_HOST_GNU_TYPE)/liblz4.so* ${BUILD_DIR}/lib || true
cp -r /lib/$(dpkg-architecture -q DEB_HOST_GNU_TYPE)/liblzma.so* ${BUILD_DIR}/lib
cp -r /lib/$(dpkg-architecture -q DEB_HOST_GNU_TYPE)/libz.so* ${BUILD_DIR}/lib

mkdir ${BUILD_DIR}/conf
cp ${DIR}/debian/snapd.service ${BUILD_DIR}/conf/
cp ${DIR}/debian/snapd.socket ${BUILD_DIR}/conf/

mkdir ${BUILD_DIR}/scripts
cp ${DIR}/tests/lib/prepare.sh ${BUILD_DIR}/scripts/
cp ${DIR}/tests/lib/apt.sh ${BUILD_DIR}/scripts/

cd ${GOPATH}

rm -rf ${NAME}-${VERSION}-${ARCH}.tar.gz
tar cpzf ${NAME}-${VERSION}-${ARCH}.tar.gz -C ${GOPATH}/build ${NAME}
