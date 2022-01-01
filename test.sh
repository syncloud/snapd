#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 2 ]; then
    echo "usage $0 version device_host"
    exit 1
fi

VERSION=$1
DEVICE_HOST=$2
apt update
apt install -y sshpass dpkg-dev
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

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"

$SCP install-snapd.sh root@${DEVICE_HOST}:/installer.sh
$SCP snapd-${VERSION}-*.tar.gz root@${DEVICE_HOST}:/

SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST}"
$SSH /installer.sh ${VERSION}

$DIR/syncloud/testapp1/build.sh
$DIR/syncloud/testapp2/build.sh
$SCP $DIR/syncloud/testapp1/testapp1.snap root@${DEVICE_HOST}:/
$SCP $DIR/syncloud/testapp2/testapp2.snap root@${DEVICE_HOST}:/

#code=0
set +e
$SSH snap install /testapp1.snap --devmode
$SSH snap install /testapp2.snap --devmode
$SSH snap install files
code=$?
#$SSH snap refresh files
#code=$(( $code + $? ))
set -e

mkdir -p log
$SSH snap changes > log/snap.changes.log   
$SSH journalctl > log/journalctl.log
exit $code
