#!/bin/bash -ex

if [[ $(lsb_release -sc) == "jessie" ]]; then
    echo "deb http://ftp.debian.org/debian jessie-backports main" > /etc/apt/sources.list.d/backports.list
fi

sudo apt-get install -y libusb-dev golang-1.6 build-essential autoconf libglib2.0-dev libseccomp-dev libapparmor-dev python-docutils libudev-dev squashfs-tools git gnupg2 gettext
rm -rf /usr/bin/go
ln -s /usr/lib/go-1.6/bin/go /usr/bin/go
rm -rf /usr/bin/gofmt
ln -s /usr/lib/go-1.6/bin/gofmt /usr/bin/gofmt