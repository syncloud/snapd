#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [[ -z "$1" ]]; then
    echo "usage $0 version"
    exit 1
fi

VERSION=$1
TESTS=$2
apt-get clean
apt update
apt install -y dpkg-dev libcap-dev libseccomp-dev xfslibs-dev
ARCH=$(dpkg-architecture -q DEB_HOST_ARCH)

NAME=snapd
BUILD_DIR=${DIR}/build/${NAME}

${DIR}/mkversion.sh ${VERSION}

if [[ ${TESTS} != "skip-tests" ]]; then
    adduser --disabled-password --gecos "" test
    chown -R test $GOPATH
    sudo -H -E -u test PATH=$PATH ${DIR}/run-checks
fi

mkdir -p ${BUILD_DIR}/bin
cp ${DIR}/.syncloud/snapd.sh ${BUILD_DIR}/bin

cd $DIR
go build -o ${BUILD_DIR}/bin/snapd github.com/snapcore/snapd/cmd/snapd
ldd ${BUILD_DIR}/bin/snapd || true
go build -ldflags '-linkmode external -extldflags -static' -tags netgo -o ${BUILD_DIR}/bin/snap github.com/snapcore/snapd/cmd/snap
go build -ldflags '-linkmode external -extldflags -static' -tags netgo -o ${BUILD_DIR}/bin/snap-exec github.com/snapcore/snapd/cmd/snap-exec
go build -ldflags '-linkmode external -extldflags -static' -tags netgo -o ${BUILD_DIR}/bin/snap-repair github.com/snapcore/snapd/cmd/snap-repair
go build -ldflags '-linkmode external -extldflags -static' -tags netgo -o ${BUILD_DIR}/bin/snap-update-ns github.com/snapcore/snapd/cmd/snap-update-ns
go build -ldflags '-linkmode external -extldflags -static' -tags netgo -o ${BUILD_DIR}/bin/snapctl github.com/snapcore/snapd/cmd/snapctl

sed -i 's/-Wl,-Bstatic//g' ${DIR}/cmd/snap-seccomp/main.go
go build -ldflags '-linkmode external -extldflags -static' -o ${BUILD_DIR}/bin/snap-seccomp github.com/snapcore/snapd/cmd/snap-seccomp

#cd  ${DIR}/cmd
#autoreconf -i -f
#./configure --disable-apparmor --disable-seccomp
#make
#cp snap-confine/snap-confine ${BUILD_DIR}/bin/snap-confine
#cp snap-discard-ns/snap-discard-ns ${BUILD_DIR}/bin/snap-discard-ns
touch ${BUILD_DIR}/bin/snap-confine
touch ${BUILD_DIR}/bin/snap-discard-ns


mkdir ${BUILD_DIR}/lib
cp -rH /usr/lib/*/libseccomp.so* ${BUILD_DIR}/lib
#cp /lib/*/ld*.so ${BUILD_DIR}/lib/ld.so
#cp /lib/*/libc*.so ${BUILD_DIR}/lib
#cp /lib/*/libpthread.so* ${BUILD_DIR}/lib

mkdir ${BUILD_DIR}/conf
cp ${DIR}/.syncloud/config/snapd.service ${BUILD_DIR}/conf/
cp ${DIR}/.syncloud/config/snapd.socket ${BUILD_DIR}/conf/

mkdir ${BUILD_DIR}/scripts
cp ${DIR}/tests/lib/prepare.sh ${BUILD_DIR}/scripts/

cp ${DIR}/.syncloud/install.sh ${BUILD_DIR}
cp ${DIR}/.syncloud/upgrade.sh ${BUILD_DIR}

cd ${DIR}

rm -rf ${NAME}-${VERSION}-${ARCH}.tar.gz
tar cpzf ${NAME}-${VERSION}-${ARCH}.tar.gz -C ${DIR}/build ${NAME}

