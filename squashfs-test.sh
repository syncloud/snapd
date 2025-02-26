#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR

./build/snapd/bin/unsquashfs -ll .syncloud/test/testapp1_1_*.snap
./build/snapd/bin/unsquashfs -h || grep
./build/snapd/bin/unsquashfs -v
./build/snapd/bin/squashfs -h || true
./build/snapd/bin/squashfs -v
