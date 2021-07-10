#!/bin/bash -ex

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

$DIR/testapp1/build.sh
$DIR/testapp2/build.sh
sshpass -p syncloud scp -o StrictHostKeyChecking=no $DIR/testapp1.snap root@${DEVICE_HOST}:/
sshpass -p syncloud scp -o StrictHostKeyChecking=no $DIR/testapp2.snap root@${DEVICE_HOST}:/

sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} snap install /testap1.snap --devmode
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} snap install /testap2.snap --devmode

mkdir log
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} snap changes > log/snap.changes.log   
sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST} journalctl > log/journalctl.log





