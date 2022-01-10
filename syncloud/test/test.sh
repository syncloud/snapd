#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ "$#" -lt 1 ]; then
    echo "usage $0 version"
    exit 1
fi

VERSION=$1
SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
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
wait_for_host apps.syncloud.org

mkdir -p $LOG_DIR

$SSH root@apps.syncloud.org apt update
$SSH root@apps.syncloud.org apt install -y nginx tree
$SSH root@apps.syncloud.org mkdir -p /var/www/html/releases/master
$SSH root@apps.syncloud.org mkdir -p /var/www/html/apps
$SSH root@apps.syncloud.org mkdir -p /var/www/html/revisions
$SCP ${DIR}/../../syncloud-release-$ARCH root@apps.syncloud.org:/syncloud-release
$SCP ${DIR}/../test/testapp1/testapp1_1_$ARCH.snap root@apps.syncloud.org:/
$SCP ${DIR}/../test/testapp2/testapp2_1_$ARCH.snap root@apps.syncloud.org:/
$SSH root@apps.syncloud.org /syncloud-release publish -f /testapp1_1_$ARCH.snap -b master -t /var/www/html
$SSH root@apps.syncloud.org /syncloud-release publish -f /testapp2_1_$ARCH.snap -b master -t /var/www/html
$SCP ${DIR}/index-v2 root@apps.syncloud.org:/var/www/html/releases/master
$SSH root@apps.syncloud.org tree /var/www/html > $LOG_DIR/store.tree.log
$SSH root@apps.syncloud.org systemctl status nginx > $LOG_DIR/nginx.status.log

$SCP ${DIR}/install-snapd.sh root@device:/installer.sh
$SCP ${DIR}/../../snapd-${VERSION}-*.tar.gz root@device:/

$SSH root@device /installer.sh ${VERSION}

#code=0
set +e
$SSH root@device snap install testapp1 --channel=master
$SSH root@device snap install testapp2 --channel=master
#$SSH snap install files
code=$?
#$SSH snap refresh files
#code=$(( $code + $? ))
set -e

#mkdir -p log
$SSH root@device snap changes > $LOG_DIR/snap.changes.log || true
$SSH root@device journalctl > $LOG_DIR/journalctl.device.log
$SSH apps.syncloud.org journalctl > $LOG_DIR/journalctl.store.log
exit $code
