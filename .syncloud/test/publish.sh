#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SCP="sshpass -p syncloud scp -o StrictHostKeyChecking=no"
SSH="sshpass -p syncloud ssh -o StrictHostKeyChecking=no"
ARTIFACTS_DIR=${DIR}/../../artifacts
mkdir -p $ARTIFACTS_DIR

apt update
apt install -y sshpass curl
cd $DIR

./wait-for-device.sh apps.syncloud.org

STORE_DIR=/var/www/html

$SSH root@apps.syncloud.org apt update
$SSH root@apps.syncloud.org apt install -y nginx tree
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/master
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/rc
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/releases/stable
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/apps
$SSH root@apps.syncloud.org mkdir -p $STORE_DIR/revisions
$SCP ${DIR}/../out/syncloud-release root@apps.syncloud.org:/

$SCP ${DIR}/testapp*.snap root@apps.syncloud.org:/

$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/master
$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/rc
$SCP ${DIR}/index-v2 root@apps.syncloud.org:$STORE_DIR/releases/stable
$SSH root@apps.syncloud.org tree $STORE_DIR > $ARTIFACTS_DIR/store.tree.log
$SSH root@apps.syncloud.org systemctl status nginx > $ARTIFACTS_DIR/nginx.status.log
