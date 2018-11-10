
apt-get -y install libncurses5

curl -so/usr/bin/zigbee-gateway https://tim-behrsin-portfolio.s3.amazonaws.com/zigbee-gateway
chmod 0755 /usr/bin/zigbee-gateway

# apt-get install libreadline-dev libncurses-dev
#
# cd /tmp
# git clone git@github.com:behrsin/gecko-sdk-2.4
# cd gecko-sdk-2.4/app/builder/ZigbeeGateway
# make
#
# cp build/exe/ZigbeeGateway /usr/bin/zigbee-gateway
# cd /tmp
# rm -Rf gecko-sdk-2.4

systemctl enable zigbee-gateway.{socket,service}
