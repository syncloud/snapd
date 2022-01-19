#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 2 ]; then
    echo "usage $0 device version"
    exit 1
fi

DEVICE=$1
VERSION=$2

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
LOG_DIR=${DIR}/../../log/$DEVICE
SNAP_ARCH=$(dpkg --print-architecture)

apt update
apt install -y sshpass curl
cd $DIR

function wait_for_host() {
  local host=$1
  attempts=100
  attempt=0
  set +e
  sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@$host date
  while test $? -gt 0
  do
    if [ $attempt -gt $attempts ]; then
      exit 1
    fi
    sleep 3
    echo "Waiting for SSH $attempt"
    attempt=$((attempt+1))
    sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@$host date
  done
  set -e
}

wait_for_host $DEVICE

mkdir -p $LOG_DIR

$SCP ${DIR}/install-snapd.sh root@$DEVICE:/installer.sh
$SCP ${DIR}/../../snapd-${VERSION}-*.tar.gz root@$DEVICE:/

set +e
$SSH root@$DEVICE /installer.sh ${VERSION}
$SCP ${DIR}/testapp2/testapp2_1_$SNAP_ARCH.snap root@$DEVICE:/testapp.snap
$SSH root@$DEVICE snap install testapp1
$SSH root@$DEVICE snap install testapp2
code=$?
set -e

$SSH root@$DEVICE snap remove testapp2 || true
$SSH root@$DEVICE snap remove testapp1 || true
$SSH root@$DEVICE snap install /testapp.snap --devmode
$SSH root@$DEVICE snap refresh testapp2 --channel=master --amend || true

$SSH root@$DEVICE snap changes > $LOG_DIR/snap.changes.log || true
$SSH root@$DEVICE journalctl > $LOG_DIR/journalctl.device.log
$SSH apps.syncloud.org journalctl > $LOG_DIR/journalctl.store.log

$SSH root@$DEVICE unsquashfs -i -d /test /testapp.snap meta/snap.yaml
$SSH root@$DEVICE ls -la /test
$SSH root@$DEVICE ls -la /test/meta
$SSH root@$DEVICE cat /test/meta/snap.yaml

exit $code
