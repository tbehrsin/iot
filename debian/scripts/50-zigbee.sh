
apt-get -y install libncurses5

curl -so/usr/bin/zigbee-gateway https://tim-behrsin-portfolio.s3.amazonaws.com/zigbee-gateway
chmod 0755 /usr/bin/zigbee-gateway

systemctl enable zigbee-gateway.{socket,service}
