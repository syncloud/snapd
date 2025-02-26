#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR

./build/snapd/bin/unsquashfs -ll .syncloud/test/testapp1_1_*.snap
./build/snapd/bin/unsquashfs -h | grep xz
./build/snapd/bin/squashfs -h | gtep xz
