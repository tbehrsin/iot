

systemctl enable iot-gateway.service
mkdir -p /data /app

rm -f /usr/bin/iot-gateway
ln -sf /app/bin/iot-gateway /usr/bin/iot-gateway
