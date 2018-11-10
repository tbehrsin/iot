
chmod 0700 /etc/ssh/authorized_keys
chmod 0600 /etc/ssh/authorized_keys/root
systemctl enable ssh

apt-get -y update
apt-get -y install vim ruby git dnsutils
gem install bundler

curl -s https://dl.google.com/go/go1.11.1.linux-armv6l.tar.gz | tar xzf - -C /usr/local

cat >>/etc/environment <<EOF
GOROOT=/usr/local/go
GOPATH=/usr/src/go
EOF

mkdir -p /usr/src/go
ln -sf /usr/local/go/bin/go /usr/local/bin/go
