#!/bin/bash

if [ -f ${GOPATH}/src/iot/vendor/github.com/augustoroman/v8/libv8 ]; then
  exit 0
fi

mkdir /tmp/libv8
cd /tmp/libv8
curl -s https://rubygems.org/downloads/libv8-6.3.292.48.1-x86_64-darwin-18.gem | tar -xvf -

tar -xzf data.tar.gz
mv $(pwd)/vendor/v8/include ${GOPATH}/src/iot/vendor/github.com/augustoroman/v8/include
mv $(pwd)/vendor/v8/out/x64.release ${GOPATH}/src/iot/vendor/github.com/augustoroman/v8/libv8

cd ${GOPATH}
rm -Rf /tmp/libv8
