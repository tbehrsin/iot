#!/bin/bash

start() {
  sleep 2
  IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' iot-server)
  if [ -z "${IP}" ]; then
    exit 0
  fi
  iptables -t nat -A POSTROUTING -s ${IP}/32 -d ${IP}/32 -p tcp -m tcp --dport 10000:65535 -j MASQUERADE
  iptables -t nat -A DOCKER ! -i docker0 -p tcp -m tcp --dport 10000:65535 -j DNAT --to-destination ${IP}:10000-65535
  iptables -A DOCKER -d ${IP}/32 ! -i docker0 -o docker0 -p tcp -m tcp --dport 10000:65535 -j ACCEPT
}

stop() {
  IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' iot-server)
  if [ -z "${IP}" ]; then
    exit 0
  fi
  iptables -t nat -D POSTROUTING -s ${IP}/32 -d ${IP}/32 -p tcp -m tcp --dport 10000:65535 -j MASQUERADE
  iptables -t nat -D DOCKER ! -i docker0 -p tcp -m tcp --dport 10000:65535 -j DNAT --to-destination ${IP}:10000-65535
  iptables -D DOCKER -d ${IP}/32 ! -i docker0 -o docker0 -p tcp -m tcp --dport 10000:65535 -j ACCEPT
}

case $1 in
  start)
    start
    ;;
  stop)
    stop
    ;;
esac
