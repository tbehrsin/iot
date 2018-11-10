#!/bin/bash -ex

apt-get -y update
apt-get -y install git ruby ruby-dev vim

gem install bundler
