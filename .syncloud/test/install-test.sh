#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SNAPD=$1
cd /tmp
rm -rf snapd
tar xzvf ${SNAPD}
sed -i 's#Environment=SNAPPY_FORCE_API_URL=.*#Environment=SNAPPY_FORCE_API_URL=http://api.store.test#g' snapd/conf/snapd.service
./snapd/install.sh
./snapd/upgrade.sh