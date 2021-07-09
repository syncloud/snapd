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

apt update
apt install -y sshpass dpkg-dev
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
$SCP snapd-${VERSION}-${ARCH}.tar.gz root@${DEVICE_HOST}:/

SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST}"

$SSH /installer.sh ${VERSION}

code=0
set +e
$SSH snap install platform
$SSH snap install files
code=$?
$SSH snap refresh files
code=$(( $code + $? ))
set -e

mkdir log
$SSH snap changes > log/snap.changes.log   
$SSH journalctl > log/journalctl.log
exit $code
