#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

app=snapd
branch=$1
build_number=$2
bucket=apps.syncloud.org
arch=$(dpkg --print-architecture)

if [ "${branch}" == "master" ] || [ "${branch}" == "stable" ] ; then
   
  s3cmd put ${app}-${build_number}-${arch}.tar.gz s3://${bucket}/apps/${app}-${build_number}-${arch}.tar.gz


fi

