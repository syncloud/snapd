#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
systemctl stop snapd.service snapd.socket || true
systemctl disable snapd.service snapd.socket || true

rm -rf /var/lib/snapd
mkdir /var/lib/snapd

rm -rf /usr/lib/snapd
mkdir -p /usr/lib/snapd
mkdir -p /var/lib/snapd/snaps
cd $DIR
cp $DIR/bin/snapd /usr/lib/snapd
cp $DIR/bin/snap-exec /usr/lib/snapd
cp $DIR/bin/snap-confine /usr/lib/snapd
cp $DIR/bin/snap-seccomp /usr/lib/snapd
cp $DIR/bin/snap-repair /usr/lib/snapd
cp $DIR/bin/snap-update-ns /usr/lib/snapd
cp $DIR/bin/snap-discard-ns /usr/lib/snapd
cp $DIR/bin/snap /usr/bin
cp $DIR/bin/snapctl /usr/bin
cp $DIR/bin/mksquashfs /usr/bin
cp $DIR/bin/unsquashfs /usr/bin
rm -rf /usr/squashfs
cp -r $DIR/squashfs /usr/

mkdir -p /usr/lib/snapd/lib
cp $DIR/lib/* /usr/lib/snapd/lib
cp $DIR/conf/snapd.service /lib/systemd/system/
cp $DIR/conf/snapd.socket /lib/systemd/system/

systemctl enable snapd.service
systemctl enable snapd.socket
systemctl start snapd.service snapd.socket

snap --version
