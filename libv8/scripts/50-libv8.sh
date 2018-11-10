#!/bin/bash -ex

cd /dist
if [ ! -d libv8 ]; then
  git clone --recursive git://github.com/cowboyd/libv8.git
  cd libv8
  git checkout v6.3.292.48.1
else
  cd libv8
fi

bundle install
bundle exec rake binary
