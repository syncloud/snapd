#!/bin/bash -e
DIR=/usr/lib/snapd
LIBS=$DIR/lib
exec $LIBS/ld.so --library-path $LIBS $DIR/snapd
