#!/bin/bash -xe
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )"  && pwd )

systemctl stop snapd.service snapd.socket || true

# TODO: /usr/lib/snapd/snapd is still busy sometimes right after the stop
sleep 5

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
mkdir -p /usr/lib/snapd/lib
cp $DIR/lib/* /usr/lib/snapd/lib

cp $DIR/conf/snapd.service /lib/systemd/system/
cp $DIR/conf/snapd.socket /lib/systemd/system/

systemctl daemon-reload
systemctl start snapd.service snapd.socket

snap --version
