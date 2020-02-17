#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 2 ]; then
    echo "usage $0 version device_host"
    exit 1
fi

VERSION=$1
DEVICE_HOST=$2
ARCH=$(dpkg-architecture -q DEB_HOST_ARCH)

cd ${DIR}

attempts=100
attempt=0

set +e
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} date
while test $? -gt 0
do
  if [ $attempt -gt $attempts ]; then
    exit 1
  fi
  sleep 3
  echo "Waiting for SSH $attempt"
  attempt=$((attempt+1))
  sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} date
done
set -e

sshpass -p syncloud scp -o StrictHostKeyChecking=no install-snapd.sh root@${DEVICE_HOST}:/installer.sh
sshpass -p syncloud scp -o StrictHostKeyChecking=no ../snapd-${VERSION}-${ARCH}.tar.gz root@${DEVICE_HOST}:/

sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} /installer.sh ${VERSION}

code=0
set +e
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} snap install files
code=$?
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} snap refresh files
code=$(( $code + $? ))
set -e

mkdir log
VERSION=$(curl http://apps.syncloud.org/releases/stable/files.version)
FILES=files_${VERSION}_${ARCH}.snap
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} LD_LIBRARY_PATH=/usr/lib/snapd/lib"
$SSH snap changes > log/snap.changes.log   
$SSH journalctl > log/journalctl.log
$SSH wget http://apps.syncloud.org/apps/${FILES} --progress=dot:giga
$SSH /usr/bin/unsquashfs --help > log/unsquashfs.log 2>&1
$SSH /usr/bin/unsquashfs -no-progress -dest . -ll $FILES
$SSH ls -la
exit $code
