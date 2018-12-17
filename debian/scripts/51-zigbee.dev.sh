
apt-get -y install libreadline-dev libtinfo-dev libncurses5-dev

cd /build
git clone git@github.com:behrsin/gecko-sdk-2.4
cd gecko-sdk-2.4/app/builder/ZigbeeGateway
make install

apt-get -y remove libreadline-dev libtinfo-dev libncurses5-dev
