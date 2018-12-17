
systemctl enable resize-partition.service

apt-get -y update
apt-get -y install wget curl git
apt-get -y remove raspi-config dhcpcd5

userdel -rf pi

mkdir -p /usr/local/share/ca-certificates/z3js/
curl -sko /usr/local/share/ca-certificates/z3js/z3js.crt http://ca.z3js.com/ca.crt
update-ca-certificates
