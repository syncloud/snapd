#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [[ -z "$1" ]]; then
    echo "usage $0 version"
    exit 1
fi

VERSION=$1
TESTS=$2
apt update
apt install -y dpkg-dev libcap-dev libseccomp-dev xfslibs-dev squashfs-tools
ARCH=$(dpkg-architecture -q DEB_HOST_ARCH)

NAME=snapd
BUILD_DIR=${DIR}/build/${NAME}

${DIR}/mkversion.sh ${VERSION}

if [[ ${TESTS} != "skip-tests" ]]; then
    adduser --disabled-password --gecos "" test
    chown -R test $GOPATH
    sudo -H -E -u test PATH=$PATH ${DIR}/run-checks
fi

rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

mkdir ${BUILD_DIR}/bin
#go get -v github.com/snapcore/snapd/...
go build -o ${BUILD_DIR}/bin/snapd github.com/snapcore/snapd/cmd/snapd
go build -o ${BUILD_DIR}/bin/snap github.com/snapcore/snapd/cmd/snap
go build -o ${BUILD_DIR}/bin/snap-exec github.com/snapcore/snapd/cmd/snap-exec
go build -o ${BUILD_DIR}/bin/snap-repair github.com/snapcore/snapd/cmd/snap-repair
go build -o ${BUILD_DIR}/bin/snap-update-ns github.com/snapcore/snapd/cmd/snap-update-ns
go build -o ${BUILD_DIR}/bin/snapctl github.com/snapcore/snapd/cmd/snapctl

sed -i 's/-Wl,-Bstatic//g' ${DIR}/cmd/snap-seccomp/main.go
go build -o ${BUILD_DIR}/bin/snap-seccomp github.com/snapcore/snapd/cmd/snap-seccomp

#cd  ${DIR}/cmd
#autoreconf -i -f
#./configure --disable-apparmor --disable-seccomp
#make
#cp snap-confine/snap-confine ${BUILD_DIR}/bin/snap-confine
#cp snap-discard-ns/snap-discard-ns ${BUILD_DIR}/bin/snap-discard-ns
touch ${BUILD_DIR}/bin/snap-confine
touch ${BUILD_DIR}/bin/snap-discard-ns

cp /usr/bin/mksquashfs ${BUILD_DIR}/bin
cp /usr/bin/unsquashfs ${BUILD_DIR}/bin

mkdir ${BUILD_DIR}/lib
cp -r /lib/*/liblzo2.so* ${BUILD_DIR}/lib
cp -r /usr/lib/*/liblz4.so* ${BUILD_DIR}/lib || true
cp -r /lib/*/liblzma.so* ${BUILD_DIR}/lib
cp -r /lib/*/libz.so* ${BUILD_DIR}/lib
cp -rH /usr/lib/*/libseccomp.so* ${BUILD_DIR}/lib


mkdir ${BUILD_DIR}/conf
cp ${DIR}/syncloud/snapd.service ${BUILD_DIR}/conf/
cp ${DIR}/syncloud/snapd.socket ${BUILD_DIR}/conf/

mkdir ${BUILD_DIR}/scripts
cp ${DIR}/tests/lib/prepare.sh ${BUILD_DIR}/scripts/

cd ${DIR}

rm -rf ${NAME}-${VERSION}-${ARCH}.tar.gz
tar cpzf ${NAME}-${VERSION}-${ARCH}.tar.gz -C ${DIR}/build ${NAME}
