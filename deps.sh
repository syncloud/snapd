#!/bin/bash

sudo apt-get install -y libusb-dev golang-1.6 build-essential autoconf libglib2.0-dev libseccomp-dev libapparmor-dev python-docutils libudev-dev squashfs-tools git gnupg2
rm -rf /usr/bin/go
ln -s /usr/lib/go-1.6/bin/go /usr/bin/go
rm -rf /usr/bin/gofmt
ln -s /usr/lib/go-1.6/bin/gofmt /usr/bin/gofmt