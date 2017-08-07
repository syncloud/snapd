#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

if [[ $(. /etc/os-release; echo $VERSION) =~ .*jessie.* ]]; then
    echo "deb http://ftp.debian.org/debian jessie-backports main" > /etc/apt/sources.list.d/backports.list
fi
apt-get update
apt-get install -y libusb-dev build-essential autoconf libglib2.0-dev libseccomp-dev libapparmor-dev python-docutils libudev-dev squashfs-tools git gnupg2 gettext xfslibs-dev libcap-dev liblz4-dev golang-1.6 wget unzip python-dateutil
rm -rf /usr/bin/go
ln -s /usr/lib/go-1.6/bin/go /usr/bin/go
rm -rf /usr/bin/gofmt
ln -s /usr/lib/go-1.6/bin/gofmt /usr/bin/gofmt

./install-s3cmd.sh
