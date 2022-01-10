#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 2 ]; then
    echo "usage $0 version device_host"
    exit 1
fi

VERSION=$1
SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@"
ARCH=$(dpkg --print-architecture)
LOG_DIR=${DIR}/../../log

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

wait_for_host device
wait_for_host store

mkdir -p $LOG_DIR

$SSH store apt update && apt install -y nginx tree
$SSH store mkdir -p /var/www/html/releases/master
$SSH store mkdir -p /var/www/html/apps
$SCP ${DIR}/../../syncloud-release-$ARCH root@store:/syncloud-release
$SCP ${DIR}/../test/testapp1/testapp1_1_$ARCH.snap root@store:/
$SCP ${DIR}/../test/testapp2/testapp2_1_$ARCH.snap root@store:/
$SSH store /syncloud-release publish -f /testapp1_1_$ARCH.snap -b master -t /var/www/html
$SSH store /syncloud-release publish -f /testapp2_1_$ARCH.snap -b master -t /var/www/html
$SCP ${DIR}/index-v2 root@store:/var/www/html/releases/master
$SSH store tree /var/www/html > $LOG_DIR/store.tree.log
$SSH store systemctl status nginx > $LOG_DIR/nginx.status.log

$SCP ${DIR}/install-snapd.sh root@device:/installer.sh
$SCP ${DIR}/../../snapd-${VERSION}-*.tar.gz root@device:/

$SSH device /installer.sh ${VERSION}

#code=0
set +e
$SSH device snap install testapp1 --channel=master
$SSH device snap install testapp2 --channel=master
#$SSH snap install files
code=$?
#$SSH snap refresh files
#code=$(( $code + $? ))
set -e

#mkdir -p log
$SSH device snap changes > $LOG_DIR/snap.changes.log || true
$SSH device journalctl > $LOG_DIR/journalctl.device.log
$SSH store journalctl > $LOG_DIR/journalctl.store.log
exit $code
