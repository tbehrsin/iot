#!/bin/bash

if [ -d ${GOPATH}/src/iot/vendor/github.com/tbehrsin/v8/libv8 ]; then
  exit 0
fi

rm -Rf /tmp/libv8
mkdir /tmp/libv8
cd /tmp/libv8
curl -sSL http://tim-behrsin-portfolio.s3.amazonaws.com/libv8-6.3.292.48.1-arm-linux.gem | tar -xvf -

tar -xzf data.tar.gz
cp -r $(pwd)/vendor/v8/include ${GOPATH}/src/iot/vendor/github.com/tbehrsin/v8/include
cp -r $(pwd)/vendor/v8/out/arm.release ${GOPATH}/src/iot/vendor/github.com/tbehrsin/v8/libv8

cd ${GOPATH}
rm -Rf /tmp/libv8
