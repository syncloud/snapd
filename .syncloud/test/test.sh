#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

STORE_DIR=/var/www/html
SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/artifacts
mkdir $ARTIFACTS_DIR
SNAP_ARCH=$(dpkg --print-architecture)

apt update
apt install -y sshpass curl wget
cd $DIR

./wait-for-device.sh device
#./wait-for-device.sh api.store.test
#./wait-for-device.sh apps.syncloud.org

#$SCP ${DIR}/../bin/install.sh root@api.store.test:/install.sh
#$SCP ${DIR}/../out/store-*.tar.gz root@api.store.test:/store.tar.gz
$SCP ${DIR}/../../snapd-*.tar.gz root@device:/snapd.tar.gz
$SCP ${DIR}/../install.sh root@device:/

#$SCP ${DIR}/testapp2_1_$SNAP_ARCH.snap root@device:/testapp2_1.snap

code=0
set +e
go test
#${DIR}/test
code=$(($code+$?))
set -e

$SSH root@device snap changes > $ARTIFACTS_DIR/snap.changes.log || true
$SSH root@device journalctl > $ARTIFACTS_DIR/journalctl.device.log
#$SCP api.store.test:/var/log/apache2/store-access.log $LOG_DIR
#$SCP api.store.test:/var/log/apache2/store-error.log $LOG_DIR
#$SSH api.store.test journalctl > $LOG_DIR/journalctl.store.log
#$SSH api.store.test ls -la /var/www/store > $LOG_DIR/var.www.store.log
#$SCP -r apps.syncloud.org:$STORE_DIR $ARTIFACTS_DIR/store
#$SCP apps.syncloud.org:/var/log/nginx/access.log $LOG_DIR/apps.nginx.access.log
chmod -R a+r $ARTIFACTS_DIR

exit $code
