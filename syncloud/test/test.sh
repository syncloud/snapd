#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 2 ]; then
    echo "usage $0 version device_host"
    exit 1
fi

VERSION=$1
DEVICE_HOST=$2
apt update
apt install -y sshpass curl
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

attempt=0
ARCH=$(dpkg --print-architecture)
until $(curl --head --fail http://apps.syncloud.org/releases/master/testapp1.$ARCH.version); do
    if [ $attempt -gt $attempts ]; then
      exit 1
    fi
    sleep 3
    attempt=$((attempt+1))
    echo "Waiting for the store $attempt"
    sleep 3
done

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"

$SCP ${DIR}/install-snapd.sh root@${DEVICE_HOST}:/installer.sh
$SCP ${DIR}/../../snapd-${VERSION}-*.tar.gz root@${DEVICE_HOST}:/

SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@${DEVICE_HOST}"
$SSH /installer.sh ${VERSION}

#code=0
set +e
$SSH snap install testapp1 --cahnnel=master
$SSH snap install testapp2 --cahnnel=master
#$SSH snap install files
code=$?
#$SSH snap refresh files
#code=$(( $code + $? ))
set -e

#mkdir -p log
$SSH snap changes > ${DIR}/../../log/snap.changes.log || true
$SSH journalctl > ${DIR}/../../log/journalctl.log
exit $code
